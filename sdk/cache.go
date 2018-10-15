package sdk

import (
	log "github.com/Sirupsen/logrus"
	"github.com/patrickmn/go-cache"
)

// readingsCache is the cache that will store the readings collected by the
// plugin, if it is enabled in the plugin configuration.
var readingsCache *cache.Cache

// cacheContexts is how ReadContexts are stored in the readings cache. Since
// we may want to filter readings based on the timestamp they were added, we
// want to store the ReadContexts against a timestamp key. In order to support
// multiple contexts at a given time, we store them as a slice.
type cacheContexts []*ReadContext

// setupReadingsCache sets up a cache that will be used to store readings,
// if it is enabled in the plugin configuration.
func setupReadingsCache() {
	cacheSettings := Config.Plugin.Settings.Cache
	if cacheSettings.Enabled {
		log.Debugf("[cache] readings cache is enabled")
		if readingsCache == nil {
			log.WithField(
				"ttl", cacheSettings.TTL,
			).Info("[cache] creating new readings cache")
			readingsCache = cache.New(cacheSettings.TTL, cacheSettings.TTL*2)
		}
	} else {
		log.Debug("[cache] readings cache disabled - will only provide current readings")
	}
}

// TODO (etd) - have a "ReadingCache" type or something to encapsulate
// all of this?

// TODO (etd) -- this needs some testing..
// addReading adds a reading to the readings cache.
func addReadingToCache(ctx *ReadContext) {
	if Config.Plugin.Settings.Cache.Enabled {
		now := GetCurrentTime()
		item, exists := readingsCache.Get(now)
		if !exists {
			newCtxs := cacheContexts([]*ReadContext{ctx})
			readingsCache.Set(now, &newCtxs, cache.DefaultExpiration)
		} else {
			cached := item.(*cacheContexts)
			*cached = append(*cached, ctx)
		}
	}
}

// todo: figure out the grpc api - we'll probably want some filtering
// on the cache and this can specify the bounds of that filter.
type readingCacheFilter struct {
	from string
	until string
}

// TODO - maybe passing in a channel and dumping all the read contexts
// to the channel makes more sense.
func getReadingsFromCache(filter *readingCacheFilter) []*ReadContext {
	var out []*ReadContext
	for _, item := range readingsCache.Items() {
		ctxs := item.Object.(*cacheContexts)
		for _, ctx := range *ctxs {
			out = append(out, ctx)
		}
	}
	return out
}