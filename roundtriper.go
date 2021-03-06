package httpcache

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gooooogolf/httpcache/cache"
)

// Headers
const (
	HeaderAuthorization = "Authorization"

	XHTTPCache          = "X-PHUB-Cache"
	XHTTPCacheOrigin    = "X-PHUB-Origin"
	XHTTPCacheExpiresIn = "X-PHUB-ExpiresIn"
	XHTTPCacheKey       = "X-PHUB-Cache-Key"
)

// CacheHandler custom plugable' struct of implementation of the http.RoundTripper
type CacheHandler struct {
	DefaultRoundTripper http.RoundTripper
	CacheInteractor     cache.ICacheInteractor
}

// NewCacheHandlerRoundtrip will create an implementations of cache http roundtripper
func NewCacheHandlerRoundtrip(defaultRoundTripper http.RoundTripper, cacheActor cache.ICacheInteractor) *CacheHandler {
	if cacheActor == nil {
		log.Fatal("cache storage is not well set")
	}
	return &CacheHandler{
		DefaultRoundTripper: defaultRoundTripper,
		CacheInteractor:     cacheActor,
	}
}

// RoundTrip the implementation of http.RoundTripper
func (r *CacheHandler) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	allowCache := allowedCache((req))
	if allowCache {
		cachedResp, cachedErr := getCachedResponse(r.CacheInteractor, req)
		if cachedResp != nil && cachedErr == nil {
			buildTheCachedResponseHeader(r.CacheInteractor, req, cachedResp)
			return cachedResp, cachedErr
		}

		// if error when getting from cache, ignore it, re-try a live version
		if cachedErr != nil {
			log.Println(cachedErr, "failed to retrieve from cache, trying with a live version")
		}
	}

	resp, err = r.DefaultRoundTripper.RoundTrip(req)
	if err != nil {
		return
	}

	if allowCache {
		err = storeRespToCache(r.CacheInteractor, req, resp)
		if err != nil {
			log.Printf("Can't store the response to database, plase check. Err: %v\n", err)
			err = nil // set err back to nil to make the call still success.
		}
	}

	return
}

func getCachedResponse(cacheInteractor cache.ICacheInteractor, req *http.Request) (resp *http.Response, err error) {
	cachedResponse, err := cacheInteractor.Get(getCacheKey(req))
	if err != nil {
		return
	}

	dumpedResponse := bytes.NewBuffer(cachedResponse)
	resp, err = http.ReadResponse(bufio.NewReader(dumpedResponse), req)
	if err != nil {
		return
	}

	return
}

func getCacheKey(req *http.Request) (key string) {
	// make sure the request URI corresponds the rewritten URL
	req.RequestURI = req.URL.Path
	if req.URL.RawQuery != "" {
		req.RequestURI = strings.Join([]string{req.RequestURI, "?", req.URL.RawQuery}, "")
	}
	key = fmt.Sprintf("%s_%s", req.Method, req.RequestURI)

	if req.Header.Get(HeaderAuthorization) != "" {
		key = fmt.Sprintf("%s_%s", key, req.Header.Get(HeaderAuthorization))
	}

	if req.Header.Get(XHTTPCacheKey) != "" {
		key = fmt.Sprintf("%s_%s", key, req.Header.Get(XHTTPCacheKey))
	}

	return
}

func storeRespToCache(cacheInteractor cache.ICacheInteractor, req *http.Request, resp *http.Response) (err error) {
	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return
	}

	err = cacheInteractor.Set(getCacheKey(req), dumpedResponse)
	return
}

// buildTheCachedResponse will finalize the response header
func buildTheCachedResponseHeader(cacheInteractor cache.ICacheInteractor, req *http.Request, resp *http.Response) {
	resp.Header.Add(XHTTPCache, "true")
	resp.Header.Add(XHTTPCacheOrigin, cacheInteractor.Origin())
	// resp.Header.Add(XHTTPCacheExpiresIn, cacheInteractor.ExpiresIn(getCacheKey(req)))
}

func allowedCache(req *http.Request) (ok bool) {
	return req.Header.Get(XHTTPCacheKey) != "" &&
		strings.Contains("GET POST", strings.ToUpper(req.Method))
}
