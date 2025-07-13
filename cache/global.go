package cache

import "time"

// Global cache instance with 30-second TTL
var AnimeSubs = NewAnimeSubsCache(30 * time.Second)
