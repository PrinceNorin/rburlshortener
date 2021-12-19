package service

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// Error return from CacheStore
var ErrCacheKeyNotFound = errors.New("cache key not found")

// CacheOption to modify caching behavior
type CacheOption struct {
	ExpiresIn time.Duration
}

// CacheStore public api
type CacheStore interface {
	Save(key string, val interface{}, opts ...CacheOption) error
	Get(key string, v interface{}) error
	Delete(key string) error
}

// NewMemoryCacheStore factory function
func NewMemoryCacheStore() CacheStore {
	return &memoryCacheStore{
		values: make(map[string]*memoryCacheValue),
	}
}

type memoryCacheValue struct {
	value     []byte
	expiredAt *time.Time
}

type memoryCacheStore struct {
	mux    sync.Mutex
	values map[string]*memoryCacheValue
}

func (c *memoryCacheStore) Save(key string, val interface{}, opts ...CacheOption) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	buf, err := json.Marshal(val)
	if err != nil {
		return err
	}

	value := memoryCacheValue{value: buf}
	if len(opts) > 0 {
		expiredAt := time.Now().Add(opts[0].ExpiresIn).UTC()
		value.expiredAt = &expiredAt
	}

	c.values[key] = &value
	return nil
}

func (c *memoryCacheStore) Get(key string, v interface{}) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	val, ok := c.values[key]
	if !ok {
		return ErrCacheKeyNotFound
	}
	if val.expiredAt != nil && val.expiredAt.Before(time.Now().UTC()) {
		delete(c.values, key)
		return ErrCacheKeyNotFound
	}
	return json.Unmarshal(val.value, v)
}

func (c *memoryCacheStore) Delete(key string) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	delete(c.values, key)
	return nil
}
