// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// AliasCache is a cache which is used to lookup devices based on any
// aliases registered by the device.
type AliasCache struct {
	// cache is a simple map which serves as the lookup table for device
	// aliases. Since device aliases will not change for the duration of
	// the plugin run (as they are statically configured), this does
	// not need any form of invalidation.
	cache map[string]*Device
}

// NewAliasCache creates a new AliasCache instance.
func NewAliasCache() *AliasCache {
	return &AliasCache{
		cache: make(map[string]*Device),
	}
}

// Add adds a device alias mapping to the cache.
func (cache *AliasCache) Add(alias string, device *Device) error {
	// If the alias already exists, return an error. It is up to the
	// configurer to ensure all aliased devices have unique aliases for
	// the plugin.
	if _, ok := cache.cache[alias]; ok {
		log.WithFields(log.Fields{
			"alias":  alias,
			"device": device.GetID(),
		}).Error("[alias] alias already exists")
		return errors.New("duplicate device alias detected")
	}

	cache.cache[alias] = device
	return nil
}

// Get gets the device associated with the specified alias. If the given
// alias is not associated with a device, this returns nil.
func (cache *AliasCache) Get(alias string) *Device {
	return cache.cache[alias]
}
