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
func getReadingsFromCache(start, end string, readings chan *ReadContext) {
	// Whether we exit the function by passing all cached readings through
	// the channel or by erroring out, we want to close the channel. This
	// will signal to caller (who provides the channel) that we are done.
	defer close(readings)

	// Parse the timestamps for the starting and ending bounds on the data
	// window, if they are set.
	startTime, err := ParseRFC3339(start)
	if err != nil {
		log.Errorf("[cache] failed to parse start time: %v", err)
	}
	endTime, err := ParseRFC3339(end)
	if err != nil {
		log.Errorf("[cache] failed to parse end time: %v", err)
	}

	// If caching reads is disabled, just return all of the current
	// tracked readings, if they fall within the specified time bound.
	// Otherwise, collect the readings from the cache.
	if Config.Plugin.Settings.Cache.Enabled {
		getCachedReadings(startTime, endTime, readings)
	} else {
		getCurrentReadings(readings)
	}
}

// getCachedReadings gets the readings from the read cache, filters them based
// on the provided start and end bounds, and passes them to the provided channel.
func getCachedReadings(start, end time.Time, readings chan *ReadContext) {
	for ts, item := range readingsCache.Items() {
		cachedTime, err := ParseRFC3339(ts)
		if err != nil {
			// If we can't parse the timestamp from the cache, an error is logged
			// and we move on. We should always be using RFC3339 formatted timestamps
			// as keys when things get inserted, so if we find something in there that
			// does not conform, it means something is wrong and we should not use it
			// (data corruption, something added incorrectly, ...)
			log.Error("[cache] failed to parse RFC3339 timestamp from cache - ignoring")
			continue
		}

		// If we have a start bound, check that the cached items are
		// within that bound. If not, ignore them.
		if !start.IsZero() && cachedTime.Before(start) {
			continue
		}

		// If we have an end bound, check that the cached items are
		// within that bound. If not, ignore them.
		if !end.IsZero() && cachedTime.After(end) {
			continue
		}

		// Pass the read contexts to the channel
		ctxs := item.Object.(*cacheContexts)
		for _, ctx := range *ctxs {
			readings <- ctx
		}
	}
}

// getCurrentReadings gets the current readings from the data manager and passes
// them to the provided channel.
func getCurrentReadings(readings chan *ReadContext) {
	for deviceID, data := range DataManager.getAllReadings() {
		// We have the device ID, but we will also want the provenance info
		// (rack, board, device), so we will need to lookup the device by ID.
		dev, ok := ctx.devices[deviceID]
		if !ok {
			log.WithField(
				"id", deviceID,
			).Error("[cache] found orphan reading (id does not match any known devices)")
			continue
		}

		readings <- &ReadContext{
			Rack:    dev.Location.Rack,
			Board:   dev.Location.Board,
			Device:  dev.id,
			Reading: data,
		}
	}
}
