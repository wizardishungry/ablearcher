package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

const checkContextModulo = 1000 // avoid the context mutex

func genKey(ctx context.Context, year [4]rune, difficultyFactor float64,
) (finalPub ed25519.PublicKey, finalPriv ed25519.PrivateKey, tries int64, finalErr error) {

	src := string(year[:])
	dst, err := hex.DecodeString(src)
	if err != nil {
		return nil, nil, tries, err
	}
	tail := append([]byte{0xED}, dst...)

	tailCmpStart := ed25519.PublicKeySize - len(tail)

	var MAX_SIG big.Int
	MAX_SIG.Exp(big.NewInt(2), big.NewInt(256), nil)
	MAX_SIG.Sub(&MAX_SIG, big.NewInt(1))
	var MAX_SIG_AS_FLOAT big.Float

	_, _, err = MAX_SIG_AS_FLOAT.Parse(MAX_SIG.String(), 10)
	if err != nil {
		return nil, nil, tries, err
	}

	MAX_SIG_CONV_TEST, _ := MAX_SIG_AS_FLOAT.Int(nil)
	if MAX_SIG.String() != MAX_SIG_CONV_TEST.String() {
		err := fmt.Errorf("converting MAX_SIG to a big.Float got a mismatch %v != %v", &MAX_SIG, MAX_SIG_CONV_TEST)
		if true {
			fmt.Println(err) // warn
		} else {
			return nil, nil, tries, err
		}
	}

	var key_threshold_as_float big.Float
	key_threshold_as_float.Mul(&MAX_SIG_AS_FLOAT, big.NewFloat(1-difficultyFactor))
	key_threshold, _ := key_threshold_as_float.Int(nil)

	group, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var mutex sync.Mutex
	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		group.Go(func() error {
			defer cancel()
			var myTries int64
			var keyAsInt big.Int
			defer func() { atomic.AddInt64(&tries, myTries) }()
			for {
				if myTries%checkContextModulo == 0 && ctx.Err() != nil {
					return ctx.Err()
				}

				pub, priv, err := ed25519.GenerateKey(rand.Reader)

				myTries++
				if err != nil {
					return err
				}

				if !bytes.Equal(pub[tailCmpStart:], tail) {
					continue
				}

				keyAsInt.SetBytes(pub)
				if keyAsInt.Cmp(key_threshold) == 1 {
					continue
				}

				mutex.Lock()
				finalPub, finalPriv = pub, priv
				mutex.Unlock()
				return nil
			}
		})
	}
	err = group.Wait()
	switch err {
	case context.Canceled:
		if finalPub != nil {
			return finalPub, finalPriv, tries, nil
		}
	}
	return finalPub, finalPriv, tries, err
}
