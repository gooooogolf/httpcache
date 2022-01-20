package cache

import (
	"errors"
)

var (
	// ErrInvalidCachedResponse will throw if the cached response is invalid
	ErrInvalidCachedResponse = errors.New("Cached Response is Invalid")
	// ErrFailedToSaveToCache will throw if the item can't be saved to cache
	ErrFailedToSaveToCache = errors.New("Failed to save item")
	// ErrCacheMissed will throw if an item can't be retrieved (due to invalid, or missing)
	ErrCacheMissed = errors.New("Cache is missing")
	// ErrStorageInternal will throw when some internal error in storage occurred
	ErrStorageInternal = errors.New("Internal error in storage")
)

// Cache storage type
const (
	CacheRedis = "REDIS"
)

// ICacheInteractor ...
type ICacheInteractor interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Flush() error
	Origin() string
}
