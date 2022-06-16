package ablearcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

const (
	gossipQueueTimeout = 10 * time.Second
	gossipChanLength   = 100
)

type gossip struct {
	group      *errgroup.Group
	sfGroup    singleflight.Group
	peers      map[string]*gossipPeer
	peersMutex sync.RWMutex
	queue      chan *storageRecordWithKey
}

type gossipPeer struct {
	queue chan *storageRecordWithKey
	items []*storageRecordWithKey
	done  chan struct{}
}

func newGossip(ctx context.Context) *gossip {
	group, ctx := errgroup.WithContext(ctx)
	g := &gossip{
		group: group,
		peers: make(map[string]*gossipPeer),
		queue: make(chan *storageRecordWithKey, gossipChanLength),
	}
	group.Go(func() error {
		return g.run(ctx)
	})
	return g
}

func (g *gossip) run(ctx context.Context) (err error) {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case next := <-g.queue:
			g.peersMutex.RLock() // TODO we should try holding the lock for more than one
			for _, peer := range g.peers {
				peer.queue <- next
			}
			g.peersMutex.RUnlock()
		}
	}
}

func (g *gossip) RemovePeer(ctx context.Context, serverURL string) (err error) {
	g.peersMutex.Lock()
	defer g.peersMutex.Unlock()
	peer, ok := g.peers[serverURL]
	if !ok {
		return fmt.Errorf("not found")
	}
	delete(g.peers, serverURL)
	close(peer.done)
	return nil
}

func (g *gossip) AddPeer(ctx context.Context, serverURL string) (err error) {
	ok := g.group.TryGo(func() error {
		_, err, _ := g.sfGroup.Do(serverURL, func() (interface{}, error) {
			g.peersMutex.Lock()
			peer := &gossipPeer{
				queue: make(chan *storageRecordWithKey, gossipChanLength),
				items: make([]*storageRecordWithKey, 0),
				done:  make(chan struct{}),
			}
			g.peers[serverURL] = peer
			g.peersMutex.Unlock()

			err := peer.run(serverURL)

			return nil, err
		})
		return err
	})
	if !ok {
		return fmt.Errorf("can't start peer")
	}
	return nil
}

func (g *gossip) Get(ctx context.Context, key storageKey) (v *storageRecord, err error) {
	return nil, fmt.Errorf("never call Get on gossip")
}

func (g *gossip) Put(ctx context.Context, key storageKey, datum *storageRecord) error {
	ctx, cancel := context.WithTimeout(ctx, gossipQueueTimeout)
	defer cancel()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case g.queue <- &storageRecordWithKey{
		storageRecord: datum,
		storageKey:    key,
	}:
		// cool
	}
	return nil
}

func (gp *gossipPeer) run(serverURL string) error {
	for {
		select {
		case <-gp.done:
			return nil
		case item := <-gp.queue:
			gp.items = append(gp.items, item)
		default:
		}
		// TODO fairness of new items arriving!

		for _, item := range gp.items {
			ctx := context.TODO()
			keyURL := serverURL + "/" + hex.EncodeToString(item.storageKey[:])
			req, err := http.NewRequestWithContext(ctx, http.MethodPut, keyURL, bytes.NewBuffer(item.body))
			if err != nil {
				return fmt.Errorf("http.NewRequestWithContext: %w", err)
			}

			setupPutRequest(req, item)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("http.DefaultClient.Do: %w", err)
			}

			resp.Body.Close()
			fmt.Println("server put status", resp.Status)
		}
		// unhappy path
		gp.items = make([]*storageRecordWithKey, 0)
	}
}
