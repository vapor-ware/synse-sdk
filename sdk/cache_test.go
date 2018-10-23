package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test setting up the readings cache when it is enabled in the config.
func Test_setupReadingsCache_Enabled(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}

	assert.Nil(t, readingsCache)
	setupReadingsCache()
	assert.NotNil(t, readingsCache)
}

// Test setting up the readings cache when it is disabled in the config.
func Test_setupReadingsCache_Disabled(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: false,
			},
		},
	}

	assert.Nil(t, readingsCache)
	setupReadingsCache()
	assert.Nil(t, readingsCache)
}

// Test adding a reading to the cache when the timestamp does not already
// exist in the cache.
// TODO: figure out how to mock out GetCurrentTime to test when the timestamp does exist
func Test_addReadingToCache_Enabled(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}
	setupReadingsCache()

	assert.Equal(t, 0, readingsCache.ItemCount())
	addReadingToCache(&ReadContext{
		Rack:    "rack",
		Board:   "board",
		Device:  "device",
		Reading: []*Reading{{Type: "test", Value: 1}},
	})
	assert.Equal(t, 1, readingsCache.ItemCount())
}

// Test adding a reading to the cache when it is disabled in the config.
func Test_addReadingToCache_Disabled(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: false,
			},
		},
	}

	assert.Nil(t, readingsCache)
	addReadingToCache(&ReadContext{
		Rack:    "rack",
		Board:   "board",
		Device:  "device",
		Reading: []*Reading{{Type: "test", Value: 1}},
	})
	assert.Nil(t, readingsCache)
}

// Test getting readings when the cache is enabled.
func Test_getReadingsFromCache_Enabled(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}
	setupReadingsCache()

	// manually add to the readingsCache
	ctx := &ReadContext{
		Rack:    "rack",
		Board:   "board",
		Device:  "device",
		Reading: []*Reading{{Type: "test", Value: 1}},
	}
	ctxs := cacheContexts([]*ReadContext{ctx})
	readingsCache.Set("2018-10-16T22:08:50.000000000Z", &ctxs, 0)

	c := make(chan *ReadContext, 10)
	go getReadingsFromCache("", "", c)

	var results []*ReadContext
	for r := range c {
		results = append(results, r)
	}
	assert.Equal(t, 1, len(results))
}

// Test getting readings from the cache with a start bound applied.
func Test_getReadingsFromCache_Enabled_Start(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}
	setupReadingsCache()

	// manually add to the readingsCache
	ctx := &ReadContext{
		Rack:    "rack",
		Board:   "board",
		Device:  "device",
		Reading: []*Reading{{Type: "test", Value: 1}},
	}
	ctxs := cacheContexts([]*ReadContext{ctx})
	readingsCache.Set("2018-10-16T22:08:50.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:51.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:52.000000000Z", &ctxs, 0)

	c := make(chan *ReadContext, 10)
	go getReadingsFromCache("2018-10-16T22:08:51.000000000Z", "", c)

	var results []*ReadContext
	for r := range c {
		results = append(results, r)
	}
	assert.Equal(t, 2, len(results))
}

// Test getting readings from the cache with an end bound applied.
func Test_getReadingsFromCache_Enabled_End(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}
	setupReadingsCache()

	// manually add to the readingsCache
	ctx := &ReadContext{
		Rack:    "rack",
		Board:   "board",
		Device:  "device",
		Reading: []*Reading{{Type: "test", Value: 1}},
	}
	ctxs := cacheContexts([]*ReadContext{ctx})
	readingsCache.Set("2018-10-16T22:08:50.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:51.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:52.000000000Z", &ctxs, 0)

	c := make(chan *ReadContext, 10)
	go getReadingsFromCache("", "2018-10-16T22:08:51.000000000Z", c)

	var results []*ReadContext
	for r := range c {
		results = append(results, r)
	}
	assert.Equal(t, 2, len(results))
}

// Test getting readings from the cache with both start and end bounds applied.
func Test_getReadingsFromCache_Enabled_StartEnd(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}
	setupReadingsCache()

	// manually add to the readingsCache
	ctx := &ReadContext{
		Rack:    "rack",
		Board:   "board",
		Device:  "device",
		Reading: []*Reading{{Type: "test", Value: 1}},
	}
	ctxs := cacheContexts([]*ReadContext{ctx})
	readingsCache.Set("2018-10-16T22:08:50.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:51.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:52.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:53.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:54.000000000Z", &ctxs, 0)

	c := make(chan *ReadContext, 10)
	go getReadingsFromCache("2018-10-16T22:08:51.500000000Z", "2018-10-16T22:08:53.000000000Z", c)
	var results []*ReadContext
	for r := range c {
		results = append(results, r)
	}
	assert.Equal(t, 2, len(results))
}

// Test getting readings from the cache when no data falls within the bounds.
func Test_getReadingsFromCache_Enabled_OutOfBounds(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}
}

// Test getting readings from the cache when the cache is disabled. This
// should lead to the current readings (the data manager state) being used.
func Test_getReadingsFromCache_Disabled(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: false,
			},
		},
	}
}

// Test getting readings when the start time is not an RFC3339-formatted
// timestamp. If this is the case, the timestamp will be ignored.
func Test_getReadingsFromCache_invalidStart(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}
	setupReadingsCache()

	// manually add to the readingsCache
	ctx := &ReadContext{
		Rack:    "rack",
		Board:   "board",
		Device:  "device",
		Reading: []*Reading{{Type: "test", Value: 1}},
	}
	ctxs := cacheContexts([]*ReadContext{ctx})
	readingsCache.Set("2018-10-16T22:08:50.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:51.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:52.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:53.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:54.000000000Z", &ctxs, 0)

	c := make(chan *ReadContext, 10)
	go getReadingsFromCache("Tues Oct 16 22:08:52 UTC 2018", "", c)
	var results []*ReadContext
	for r := range c {
		results = append(results, r)
	}
	assert.Equal(t, 5, len(results))
}

// Test getting readings when the end time is not an RFC3339-formatted
// timestamp. If this is the case, the timestamp will be ignored.
func Test_getReadingsFromCache_invalidEnd(t *testing.T) {
	defer func() {
		// reset plugin state
		resetContext()
		Config.reset()

		// reset readings cache
		readingsCache = nil
	}()

	Config.Plugin = &PluginConfig{
		Settings: &PluginSettings{
			Cache: &CacheSettings{
				Enabled: true,
			},
		},
	}
	setupReadingsCache()

	// manually add to the readingsCache
	ctx := &ReadContext{
		Rack:    "rack",
		Board:   "board",
		Device:  "device",
		Reading: []*Reading{{Type: "test", Value: 1}},
	}
	ctxs := cacheContexts([]*ReadContext{ctx})
	readingsCache.Set("2018-10-16T22:08:50.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:51.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:52.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:53.000000000Z", &ctxs, 0)
	readingsCache.Set("2018-10-16T22:08:54.000000000Z", &ctxs, 0)

	c := make(chan *ReadContext, 10)
	go getReadingsFromCache("", "Tues Oct 16 22:08:53 UTC 2018", c)
	var results []*ReadContext
	for r := range c {
		results = append(results, r)
	}
	assert.Equal(t, 5, len(results))
}
