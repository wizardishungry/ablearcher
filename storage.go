package ablearcher

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"
)

type storageProvider interface {
	Get(ctx context.Context, key storageKey) (*storageRecord, error)
	Put(ctx context.Context, key storageKey, datum *storageRecord) error

	AddPeer(ctx context.Context, serverURL string) error
	RemovePeer(ctx context.Context, serverURL string) error
}

type storageRecord struct {
	body []byte
	time time.Time
	auth []byte
}

type storageRecordWithKey struct {
	*storageRecord
	storageKey
}

var (
	_ storageProvider = (*memoryStore)(nil)
	_ storageProvider = (*chainedStorage)(nil)
	_ storageProvider = (*gossip)(nil)
)

type storageKey [ed25519.PublicKeySize]byte

func sliceToStorageKey(bs []byte) (s storageKey, err error) {
	if len(bs) != len(s) {
		return s, fmt.Errorf("key length does not match storageKey length (%d != %d)", len(bs), len(s))
	}
	return *(*storageKey)(bs), nil
}

var (
	_ map[storageKey]*storageRecord
)

type memoryStore struct {
	data sync.Map // map[string]*storageRecord
}

func (ms *memoryStore) Get(ctx context.Context, key storageKey) (v *storageRecord, err error) {
	datum, ok := ms.data.Load(key)
	if !ok {
		return v, fmt.Errorf("not found")
	}
	return datum.(*storageRecord), nil
}

func (ms *memoryStore) Put(ctx context.Context, key storageKey, datum *storageRecord) error {
	ms.data.Store(key, datum)
	return nil
}
func (ms *memoryStore) RemovePeer(ctx context.Context, serverURL string) error {
	return nil
}
func (ms *memoryStore) AddPeer(ctx context.Context, serverURL string) error {
	return nil
}

type chainedStorage []storageProvider

var ErrNoStorageProviders = fmt.Errorf("no storage providers")

func (cs chainedStorage) Get(ctx context.Context, key storageKey) (v *storageRecord, err error) {
	// only uses the first provider
	if len(cs) < 1 {
		return nil, ErrNoStorageProviders
	}
	return cs[0].Get(ctx, key)
}
func (cs chainedStorage) Put(ctx context.Context, key storageKey, datum *storageRecord) error {
	for _, p := range cs {
		err := p.Put(ctx, key, datum)
		if err != nil {
			return err
		}
	}
	return nil
}
func (cs chainedStorage) RemovePeer(ctx context.Context, serverURL string) error {
	for _, p := range cs {
		err := p.RemovePeer(ctx, serverURL)
		if err != nil {
			return err
		}
	}
	return nil
}
func (cs chainedStorage) AddPeer(ctx context.Context, serverURL string) error {
	for _, p := range cs {
		err := p.AddPeer(ctx, serverURL)
		if err != nil {
			return err
		}
	}
	return nil
}
