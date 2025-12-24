package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aqstack/mimir/pkg/api"
)

// MemoryCache implements an in-memory semantic cache.
type MemoryCache struct {
	mu      sync.RWMutex
	entries []*api.CacheEntry
	opts    *Options

	// Stats
	hits   atomic.Int64
	misses atomic.Int64
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache(opts *Options) *MemoryCache {
	if opts == nil {
		opts = DefaultOptions()
	}

	mc := &MemoryCache{
		entries: make([]*api.CacheEntry, 0, opts.MaxSize),
		opts:    opts,
	}

	// Start cleanup goroutine
	go mc.cleanupLoop()

	return mc
}

// Get retrieves a cached response based on semantic similarity.
func (m *MemoryCache) Get(ctx context.Context, embedding []float64, threshold float64) (*api.CacheEntry, float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var bestMatch *api.CacheEntry
	var bestSimilarity float64

	now := time.Now()

	for _, entry := range m.entries {
		// Skip expired entries
		if now.After(entry.ExpiresAt) {
			continue
		}

		similarity := CosineSimilarity(embedding, entry.Embedding)
		if similarity >= threshold && similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMatch = entry
		}
	}

	if bestMatch != nil {
		m.hits.Add(1)
		// Update hit stats (requires write lock, but we defer to avoid complexity)
		go m.updateHitStats(bestMatch)
		return bestMatch, bestSimilarity, true
	}

	m.misses.Add(1)
	return nil, 0, false
}

// updateHitStats updates the hit statistics for an entry.
func (m *MemoryCache) updateHitStats(entry *api.CacheEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry.HitCount++
	entry.LastHitAt = time.Now()
}

// Set stores a response with its embedding.
func (m *MemoryCache) Set(ctx context.Context, entry *api.CacheEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate (update if exists)
	for i, e := range m.entries {
		similarity := CosineSimilarity(entry.Embedding, e.Embedding)
		if similarity > 0.99 {
			// Update existing entry
			m.entries[i] = entry
			return nil
		}
	}

	// Evict if at capacity (LRU-style: remove oldest)
	if len(m.entries) >= m.opts.MaxSize {
		m.evictOldest()
	}

	m.entries = append(m.entries, entry)
	return nil
}

// evictOldest removes the oldest entry based on last hit time.
func (m *MemoryCache) evictOldest() {
	if len(m.entries) == 0 {
		return
	}

	oldestIdx := 0
	oldestTime := m.entries[0].LastHitAt

	for i, e := range m.entries {
		if e.LastHitAt.Before(oldestTime) {
			oldestIdx = i
			oldestTime = e.LastHitAt
		}
	}

	// Remove by swapping with last element
	m.entries[oldestIdx] = m.entries[len(m.entries)-1]
	m.entries = m.entries[:len(m.entries)-1]
}

// Delete removes an entry by its embedding.
func (m *MemoryCache) Delete(ctx context.Context, embedding []float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, e := range m.entries {
		similarity := CosineSimilarity(embedding, e.Embedding)
		if similarity > 0.99 {
			m.entries[i] = m.entries[len(m.entries)-1]
			m.entries = m.entries[:len(m.entries)-1]
			return nil
		}
	}

	return nil
}

// Clear removes all entries from the cache.
func (m *MemoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make([]*api.CacheEntry, 0, m.opts.MaxSize)
	m.hits.Store(0)
	m.misses.Store(0)

	return nil
}

// Stats returns cache statistics.
func (m *MemoryCache) Stats(ctx context.Context) *api.CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hits := m.hits.Load()
	misses := m.misses.Load()
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	// Estimate cost savings (rough: $0.002 per 1K tokens, assume 500 tokens per request)
	estimatedSaved := float64(hits) * 0.001

	return &api.CacheStats{
		TotalEntries:   int64(len(m.entries)),
		TotalHits:      hits,
		TotalMisses:    misses,
		HitRate:        hitRate,
		EstimatedSaved: estimatedSaved,
	}
}

// Cleanup removes expired entries.
func (m *MemoryCache) Cleanup(ctx context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	removed := 0

	// Filter out expired entries
	active := make([]*api.CacheEntry, 0, len(m.entries))
	for _, e := range m.entries {
		if now.Before(e.ExpiresAt) {
			active = append(active, e)
		} else {
			removed++
		}
	}

	m.entries = active
	return removed
}

// Size returns the number of entries in the cache.
func (m *MemoryCache) Size(ctx context.Context) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// cleanupLoop periodically removes expired entries.
func (m *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(m.opts.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.Cleanup(context.Background())
	}
}
