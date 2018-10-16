package sdk

import (
	"time"

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

// getReadingsFromCache takes optional start/end bounds (which can be left as empty strings
// to specify no bounds on the reading data) and a channel. It will collect all pertinent
// reading data from the cache and pass it through the channel. Once the function returns,
// the channel will be closed.
func getReadingsFromCache(start, end string, readings chan *ReadContext) { // nolint: gocyclo

	// TODO (etd): If cached readings are disabled for the plugin, just
	// get and return the current readings.

	// Whether we exit the function by passing all cached readings through
	// the channel or by erroring out, we want to close the channel. This
	// will signal to caller (who provides the channel) that we are done.
	defer close(readings)
	var err error

	// Parse the timestamp for the starting bound on the data window,
	// if it is set.
	var startTime time.Time
	if start != "" {
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			log.WithField(
				"time", start,
			).Error("[cache] failed to parse start time to RFC3339")
		}
	}

	// Parse the timestamp for the ending bound on the data window,
	// if it is set.
	var endTime time.Time
	if end != "" {
		endTime, err = time.Parse(time.RFC3339, end)
		if err != nil {
			log.WithField(
				"time", end,
			).Error("[cache] failed to parse end time to RFC3339")
		}
	}

	// Collect all of the cached readings. If starting or ending bounds
	// are specified, ensure that the data is within those bounds before
	// passing it back to the caller via the channel.

	for ts, item := range readingsCache.Items() {
		cachedTime, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			log.WithField(
				"timestamp", ts,
			).Error("[cache] failed to parse timestamp from cache")
		}

		// If we have a start bound, check that the cached items are
		// within that bound. If not, ignore them.
		if !startTime.IsZero() {
			if cachedTime.Before(startTime) {
				continue
			}
		}
		// If we have an end bound, check that the cached items are
		// within that bound. If not, ignore them.
		if !endTime.IsZero() {
			if cachedTime.After(endTime) {
				continue
			}
		}

		ctxs := item.Object.(*cacheContexts)
		for _, ctx := range *ctxs {
			readings <- ctx
		}
	}
}
