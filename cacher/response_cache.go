package cacher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-midway/midway"
	"github.com/go-midway/midway/logcontext"
	rcache "gopkg.in/go-redis/cache.v5"
	redis "gopkg.in/redis.v5"
)

// Error represents error in httpcache
type Error int

const (
	// HeaderNotExists represents error if header field is empty
	// or does not exists
	HeaderNotExists Error = iota
)

func (err Error) Error() string {
	if err == HeaderNotExists {
		return "header field not exits"
	}
	return "unknown error"
}

type rediser interface {
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
	Del(keys ...string) *redis.IntCmd
}

// HKT stores *time.Location of Hong Kong
var HKT *time.Location

func init() {
	HKT, _ = time.LoadLocation("Asia/Hong_Kong")
}

const fmtRFC2612 = "Mon, 02 Jan 2006 15:04:05 GMT"

func parseRFC2612(str string) (t time.Time, err error) {
	if len(str) < 3 {
		err = fmt.Errorf("incorrect time string provided: %s", str)
		return
	}
	t, err = time.Parse(fmtRFC2612, str[:len(str)-3]+"GMT")
	if err != nil {
		// undo the string replacenemt for better error messages
		terr := err.(*time.ParseError)
		err = fmt.Errorf("cannot parse time %#v as RFC2612 (%#v)", str, terr.Layout)
	}
	return
}

// NewResponseCache wraps an http.ResponswWriter with Cache
func NewResponseCache(w http.ResponseWriter) *ResponseCache {
	return &ResponseCache{
		responseWriter: w,
		Created:        time.Now(),
		Status:         http.StatusOK,
		content:        bytes.NewBuffer(make([]byte, 0, 4096)), // pre=alloc 4096 bytes for buffer
	}
}

// ResponseCache wraps an http.ResponseWriter and return
type ResponseCache struct {
	responseWriter http.ResponseWriter
	content        *bytes.Buffer
	Created        time.Time
	Status         int
	CachedHeader   http.Header // header loaded from cache
	CachedContent  string
}

// Code returns the cached http status code
func (cache *ResponseCache) Code() int {
	return cache.Status // TODO; ensure the default value is http.StatusOK
}

// TODO: add method to check Last-Modified date / Date of the cached response

func parseTime(header http.Header, name string) (parsed time.Time, err error) {
	var timeStr string
	if timeStr = header.Get(name); timeStr == "" {
		err = HeaderNotExists
		return
	}
	return parseRFC2612(timeStr)
}

// String return buffered content as string
func (cache *ResponseCache) String() string {
	if cache.responseWriter != nil {
		return cache.content.String()
	}
	return cache.CachedContent
}

// Bytes return buffered content as string
func (cache *ResponseCache) Bytes() []byte {
	if cache.responseWriter != nil {
		return cache.content.Bytes()
	}
	return []byte(cache.CachedContent)
}

// WriteTo writes the content of current cache to http.ResponseWriter
func (cache *ResponseCache) WriteTo(w http.ResponseWriter) {
	header := cache.Header()
	for name := range header {
		for i := range header[name] {
			w.Header().Add(name, header[name][i])
		}
	}
	w.WriteHeader(cache.Code())
	w.Write(cache.Bytes())
}

// Header implements http.ResponseWriter
func (cache *ResponseCache) Header() http.Header {
	if cache == nil {
		return nil
	}
	if cache.responseWriter != nil {
		return cache.responseWriter.Header()
	}
	return cache.CachedHeader
}

// Write implements http.ResponseWriter
func (cache *ResponseCache) Write(p []byte) (int, error) {
	cache.content.Write(p) // omit write number and error in buffer write
	return cache.responseWriter.Write(p)
}

// WriteHeader implements http.ResponseWriter
func (cache *ResponseCache) WriteHeader(code int) {
	cache.Status = code
	cache.responseWriter.WriteHeader(code)
}

func keyOf(r *http.Request) (key string, err error) {
	if r == nil {
		err = fmt.Errorf("request cannot be nil")
		return
	}
	if r.URL == nil {
		err = fmt.Errorf("r.URL cannot be nil")
		return
	}
	key = "page:/" + r.URL.Path
	return
}

// Cacher handles the response cache for the CachedHandler
type Cacher struct {
	redisCache *rcache.Codec
}

// NewCacher returns a *ResponseCacher of the given redis options
func NewCacher(opts *redis.Options) *Cacher {
	return &Cacher{
		redisCache: &rcache.Codec{
			Redis: redis.NewRing(&redis.RingOptions{
				Addrs: map[string]string{
					"default": opts.Addr,
				},
				Password: opts.Password,
			}),
			Marshal: func(v interface{}) ([]byte, error) {
				return json.Marshal(v)
			},
			Unmarshal: func(b []byte, v interface{}) error {
				return json.Unmarshal(b, v)
			},
		},
	}
}

// LoadResponse cache for a given http request
func (rcacher Cacher) LoadResponse(r *http.Request) (cache *ResponseCache, err error) {
	key, err := keyOf(r)
	if err != nil {
		return
	}

	if rcacher.redisCache == nil {
		return
	}

	cache = &ResponseCache{}
	if err = rcacher.redisCache.Get(key, cache); err != nil {
		cache = nil
		return
	}
	return
}

// SaveResponse cache for a given http request
func (rcacher Cacher) SaveResponse(r *http.Request, cache *ResponseCache) (err error) {
	key, err := keyOf(r)
	if err != nil {
		return
	}

	if rcacher.redisCache == nil {
		return
	}

	// ensure that the header is cached
	cache.CachedHeader = cache.Header()
	cache.CachedContent = cache.String()

	// store the httpcache item in memcached
	return rcacher.redisCache.Set(&rcache.Item{
		Key:        key,
		Object:     cache,
		Expiration: 60 * time.Minute, // TODO: detect correct expiration time
	})
}

// DeleteResponse deletes ResponseCache of a given request
func (rcacher Cacher) DeleteResponse(r *http.Request) (err error) {
	key, err := keyOf(r)
	if err != nil {
		return
	}

	if rcacher.redisCache == nil {
		return
	}

	return rcacher.redisCache.Delete(key)
}

// Valid test if a cache has valid cache
func Valid(r *http.Request, cache *ResponseCache) bool {
	var expires time.Time
	var err error
	logger := logcontext.GetComplexLogger(r.Context())

	// TODO: might support max-age somehow?

	// parse grace expires override
	if expires, err = parseTime(cache.Header(), "X-Grace-Expires"); err == nil {
		if expires.After(time.Now()) {
			logger.Log("message", "cache graced")
			return true
		}
	} else if err != HeaderNotExists {
		logger.Error("message", fmt.Sprintf("error parsing X-Grace-Expires (%s)", err.Error()))
	}

	if expires, err = parseTime(cache.Header(), "Expires"); err == nil {
		if expires.After(time.Now()) {
			logger.Log("message", "cache not expired")
			return true
		}
	} else if err != HeaderNotExists {
		logger.Error("message", fmt.Sprintf("error parsing Expires (%s)", err.Error()))
	}
	return false // default treat as expired
}

// CachedHandler applies httpcache to the wrapped http.Handler
func CachedHandler(rcacher *Cacher) midway.Middleware {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			logger := logcontext.GetComplexLogger(r.Context())

			// try to load cache for the request
			cache, err := rcacher.LoadResponse(r)
			if err != nil {
				logger.Error("message", fmt.Sprintf("error loading cache: %s", err.Error()))
			}

			// if has cache, write to ResponseWriter and return early
			if Valid(r, cache) {
				logger.Log("message", "use cache")
				cache.WriteTo(w)
				return // early return
			}

			// refresh cache by running inner handler
			logger.Log("message", "no valid cache, trigger inner handler")

			cache = NewResponseCache(w)
			inner.ServeHTTP(cache, r)
			go func() {
				err := rcacher.SaveResponse(r, cache)
				if err != nil {
					logger.Error("message", fmt.Sprintf("error saving cache: %s", err.Error()))
				}
			}()
		})
	}
}