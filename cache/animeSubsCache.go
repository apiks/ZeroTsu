package cache

import (
	"log"
	"time"

	"github.com/r-anime/ZeroTsu/entities"
	"github.com/sasha-s/go-deadlock"
)

// AnimeSubsCache provides thread-safe caching for anime subscriptions with TTL
type AnimeSubsCache struct {
	deadlock.RWMutex
	data      map[string][]*entities.ShowSub
	lastFetch time.Time
	ttl       time.Duration
}

// NewAnimeSubsCache creates a new anime subscriptions cache with the specified TTL
func NewAnimeSubsCache(ttl time.Duration) *AnimeSubsCache {
	return &AnimeSubsCache{
		data: make(map[string][]*entities.ShowSub),
		ttl:  ttl,
	}
}

// Get retrieves all anime subscriptions, refreshing the cache if needed
func (c *AnimeSubsCache) Get() map[string][]*entities.ShowSub {
	c.RLock()
	if time.Since(c.lastFetch) < c.ttl {
		defer c.RUnlock()
		return c.data
	}
	c.RUnlock()

	// Refresh cache
	c.Lock()
	defer c.Unlock()

	// Double-check after acquiring write lock
	if time.Since(c.lastFetch) < c.ttl {
		return c.data
	}

	animeSubs, err := entities.LoadAnimeSubs()
	if err != nil {
		log.Println("Error loading all anime subscriptions:", err)
		return c.data // Return stale data if fetch fails
	}

	c.data = animeSubs
	c.lastFetch = time.Now()
	return c.data
}

// GetGuild retrieves only guild anime subscriptions, refreshing the cache if needed
func (c *AnimeSubsCache) GetGuild() map[string][]*entities.ShowSub {
	allSubs := c.Get()

	c.RLock()
	defer c.RUnlock()

	// Count guild subscriptions first to pre-allocate map
	guildCount := 0
	for _, subs := range allSubs {
		if len(subs) > 0 && subs[0].GetGuild() {
			guildCount++
		}
	}

	guildSubs := make(map[string][]*entities.ShowSub, guildCount)
	for id, subs := range allSubs {
		if len(subs) > 0 && subs[0].GetGuild() {
			guildSubs[id] = subs
		}
	}

	return guildSubs
}

// GetUser retrieves only user anime subscriptions, refreshing the cache if needed
func (c *AnimeSubsCache) GetUser() map[string][]*entities.ShowSub {
	allSubs := c.Get()

	c.RLock()
	defer c.RUnlock()

	// Count user subscriptions first to pre-allocate map
	userCount := 0
	for _, subs := range allSubs {
		if len(subs) > 0 && !subs[0].GetGuild() {
			userCount++
		}
	}

	userSubs := make(map[string][]*entities.ShowSub, userCount)
	for id, subs := range allSubs {
		if len(subs) > 0 && !subs[0].GetGuild() {
			userSubs[id] = subs
		}
	}

	return userSubs
}

// GetByID retrieves a specific user's or guild's anime subscriptions
func (c *AnimeSubsCache) GetByID(id string) []*entities.ShowSub {
	allSubs := c.Get()

	c.RLock()
	defer c.RUnlock()

	if subs, exists := allSubs[id]; exists {
		return subs
	}
	return []*entities.ShowSub{}
}

// Invalidate forces a cache refresh on the next Get() call
func (c *AnimeSubsCache) Invalidate() {
	c.Lock()
	defer c.Unlock()
	c.lastFetch = time.Time{} // Zero time forces refresh
}

// Update updates a specific user's or guild's subscriptions in the cache
func (c *AnimeSubsCache) Update(id string, subscriptions []*entities.ShowSub) {
	c.Lock()
	defer c.Unlock()
	c.data[id] = subscriptions
}

// UpdateMultiple updates multiple users' or guilds' subscriptions in the cache
func (c *AnimeSubsCache) UpdateMultiple(updates map[string][]*entities.ShowSub) {
	c.Lock()
	defer c.Unlock()
	for id, subscriptions := range updates {
		c.data[id] = subscriptions
	}
}

// Remove removes a specific user's or guild's subscriptions from the cache
func (c *AnimeSubsCache) Remove(id string) {
	c.Lock()
	defer c.Unlock()
	delete(c.data, id)
}

// Clear clears all cached data
func (c *AnimeSubsCache) Clear() {
	c.Lock()
	defer c.Unlock()
	c.data = make(map[string][]*entities.ShowSub)
	c.lastFetch = time.Time{}
}

// IsStale checks if the cache is stale (needs refresh)
func (c *AnimeSubsCache) IsStale() bool {
	c.RLock()
	defer c.RUnlock()
	return time.Since(c.lastFetch) >= c.ttl
}

// GetLastFetch returns the last fetch time
func (c *AnimeSubsCache) GetLastFetch() time.Time {
	c.RLock()
	defer c.RUnlock()
	return c.lastFetch
}

// GetCacheSize returns the number of cached entries
func (c *AnimeSubsCache) GetCacheSize() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.data)
}
