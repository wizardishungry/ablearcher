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
}

type storageRecord struct {
	body []byte
	time time.Time
}

var (
	_ storageProvider = (*memoryStore)(nil)
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

type memoryStore = memoryStoreImpl[*storageRecord]
type memoryStoreImpl[T any] struct {
	data sync.Map // map[string][]byte
}

func (ms *memoryStoreImpl[T]) Get(ctx context.Context, key storageKey) (v T, err error) {
	datum, ok := ms.data.Load(key)
	if !ok {
		return v, fmt.Errorf("not found")
	}
	return datum.(T), nil
}

func (ms *memoryStoreImpl[T]) Put(ctx context.Context, key storageKey, datum T) error {
	ms.data.Store(key, datum)
	return nil
}
