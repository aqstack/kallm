package cache

import (
	"context"
	"time"

	"github.com/aqstack/mimir/pkg/api"
)

// Cache defines the interface for semantic caching.
type Cache interface {
	// Get retrieves a cached response based on semantic similarity.
	// Returns the cached response, similarity score, and whether a match was found.
	Get(ctx context.Context, embedding []float64, threshold float64) (*api.CacheEntry, float64, bool)

	// Set stores a response with its embedding.
	Set(ctx context.Context, entry *api.CacheEntry) error

	// Delete removes an entry by its embedding.
	Delete(ctx context.Context, embedding []float64) error

	// Clear removes all entries from the cache.
	Clear(ctx context.Context) error

	// Stats returns cache statistics.
	Stats(ctx context.Context) *api.CacheStats

	// Cleanup removes expired entries.
	Cleanup(ctx context.Context) int

	// Size returns the number of entries in the cache.
	Size(ctx context.Context) int
}

// SearchResult represents a cache search result.
type SearchResult struct {
	Entry      *api.CacheEntry
	Similarity float64
}

// Options configures cache behavior.
type Options struct {
	MaxSize             int
	DefaultTTL          time.Duration
	CleanupInterval     time.Duration
	SimilarityThreshold float64
}

// DefaultOptions returns sensible defaults for cache options.
func DefaultOptions() *Options {
	return &Options{
		MaxSize:             10000,
		DefaultTTL:          24 * time.Hour,
		CleanupInterval:     5 * time.Minute,
		SimilarityThreshold: 0.95,
	}
}
