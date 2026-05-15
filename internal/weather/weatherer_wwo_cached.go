// weather/cached_client.go
package weather

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/chubin/wttr.in/internal/domain"
)

// MakeWeatherCacheKey generates a normalized, stable cache key for weather requests.
// Lat/lon are rounded to ~100m precision to avoid cache fragmentation.
func MakeWeatherCacheKey(lat, lon float64, lang string) string {
	// Round to 4 decimal places (~11 meter precision)
	lat = math.Round(lat*10000) / 10000
	lon = math.Round(lon*10000) / 10000

	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = "en"
	}

	return fmt.Sprintf("weather:wwo:%.4f:%.4f:%s", lat, lon, lang)
}

// CachedWeatherClient wraps the original WeatherClient with disk/memory caching
// and request coalescing.
type CachedWeatherClient struct {
	client     *WeatherClient
	cacher     Cacher
	defaultTTL time.Duration
}

// NewCachedWeatherClient creates a new caching wrapper around a WeatherClient.
func NewCachedWeatherClient(
	client *WeatherClient,
	cacher Cacher,
	defaultTTL time.Duration,
) *CachedWeatherClient {
	if defaultTTL <= 0 {
		defaultTTL = 45 * time.Minute // sensible default for weather data
	}

	return &CachedWeatherClient{
		client:     client,
		cacher:     cacher,
		defaultTTL: defaultTTL,
	}
}

// GetWeather returns weather data, using cache when possible.
func (c *CachedWeatherClient) GetWeather(lat, lon float64, lang string) ([]byte, error) {
	key := MakeWeatherCacheKey(lat, lon, lang)

	// 1. Fast path: check cache
	if entry := c.cacher.Get(key); entry != nil {
		return entry.Body, nil
	}

	// 2. Try to become the leader (coalescing)
	if c.cacher.SetInProgressIfNotExists(key) {
		// We are the one fetching
		return c.fetchAndCache(key, lat, lon, lang)
	}

	// 3. Another goroutine is already fetching → wait for it
	return c.waitForInProgress(key)
}

// fetchAndCache performs the actual upstream request and stores the result.
func (c *CachedWeatherClient) fetchAndCache(key string, lat, lon float64, lang string) ([]byte, error) {
	// Ensure we always clean up the in-progress state
	defer func() {
		if r := recover(); r != nil {
			c.cacher.Remove(key)
			panic(r)
		}
	}()

	body, err := c.client.GetWeather(lat, lon, lang)
	if err != nil {
		c.cacher.Remove(key) // do not cache errors
		return nil, fmt.Errorf("weather upstream error: %w", err)
	}

	entry := domain.CacheEntry{
		Body:       body,
		StatusCode: http.StatusOK,
		Header:     make(http.Header), // extend with real headers if needed

		Expires:  time.Now().Add(c.defaultTTL),
		CachedAt: time.Now(),
		TTL:      c.defaultTTL,
		Key:      key,
		Source:   "weather:wwo",
	}

	c.cacher.Set(key, entry)
	return body, nil
}

// waitForInProgress waits for another goroutine to complete the request.
func (c *CachedWeatherClient) waitForInProgress(key string) ([]byte, error) {
	entry, err := c.cacher.WaitForCompletion(key, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("timeout waiting for weather data: %w", err)
	}
	if entry == nil {
		return nil, fmt.Errorf("weather request failed upstream (removed from cache)")
	}

	return entry.Body, nil
}

// Close closes both the underlying client and the cache.
func (c *CachedWeatherClient) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	if c.cacher != nil {
		return c.cacher.Close()
	}
	return nil
}
