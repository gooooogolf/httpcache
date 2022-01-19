package httpcache

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/gooooogolf/httpcache/cache"

	rediscache "github.com/gooooogolf/httpcache/cache/redis"
)

func newClient(client *http.Client, cacheInteractor cache.ICacheInteractor) (cachedHandler *CacheHandler, err error) {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	cachedHandler = NewCacheHandlerRoundtrip(client.Transport, cacheInteractor)
	client.Transport = cachedHandler
	return
}

// NewWithRedisCache will create a complete cache-support of HTTP client with using redis cache.
func NewWithRedisCache(client *http.Client, options *rediscache.CacheOptions,
	duration ...time.Duration) (cachedHandler *CacheHandler, err error) {
	var ctx = context.Background()
	var expiryTime time.Duration
	if len(duration) > 0 {
		expiryTime = duration[0]
	}
	c := redis.NewClient(&redis.Options{
		Addr:     options.Addr,
		Password: options.Password,
		DB:       options.DB,
	})

	return newClient(client, rediscache.NewCache(ctx, c, expiryTime))
}
