// Synse SDK
// Copyright (c) 2019 Vapor IO
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
	"os"
	"strings"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// pluginID is used to generate and store the plugin ID namespace. After
// initialized, it can be used to generate IDs within the plugin namespace
// for devices.
type pluginID struct {
	config     *config.IDSettings
	components []string
	name       string
	uuid       uuid.UUID
}

// newPluginID creates a new instance of a pluginID.
func newPluginID(conf *config.IDSettings, meta *PluginMetadata) (*pluginID, error) {
	if conf == nil {
		return nil, errors.New("unable to create plugin id: nil config")
	}
	if meta == nil {
		return nil, errors.New("unable to create plugin id: nil metadata")
	}

	var components []string

	// Add the plugin metadata tag as a component.
	if conf.UsePluginTag {
		log.WithFields(log.Fields{
			"tag": meta.Tag(),
		}).Debug("[id] using plugin tag as uuid component")
		components = append(components, meta.Tag())
	}

	// Add the machine ID as a component.
	if conf.UseMachineID {
		id, err := machineid.ProtectedID(meta.Tag())
		if err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{
			"machineID": id,
		}).Debug("[id] using machine id as uuid component")
		components = append(components, id)
	}

	// Add environment variables as a component.
	if len(conf.UseEnv) > 0 {
		for _, k := range conf.UseEnv {
			val, found := os.LookupEnv(k)
			if !found {
				return nil, errors.New("unable to create plugin id: env enabled but not set")
			}
			log.WithFields(log.Fields{
				"envVar": k,
			}).Debug("[id] using env variable tag as uuid component")
			components = append(components, val)
		}
	}

	// Add custom identifiers as a component.
	if len(conf.UseCustom) > 0 {
		log.Debug("[id] using custom component in uuid")
		components = append(components, conf.UseCustom...)
	}

	// If there are no namespace components, we are not able to generate an ID.
	if len(components) == 0 {
		return nil, errors.New("unable to create plugin id: no namespace components")
	}

	// Generate the V5 UUID for the plugin. The various ID components are joined
	// into a single string and used as the name. ORDER MATTERS.
	name := strings.Join(components, ".")
	pluginUUID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(name))

	log.WithFields(log.Fields{
		"id": pluginUUID,
	}).Info("[id] generated plugin id namespace")
	return &pluginID{
		config:     conf,
		components: components,
		name:       name,
		uuid:       pluginUUID,
	}, nil
}

// NewNamespacedID generates a new UUID based off of the pluginID's namespaced
// ID. This function should be used to generate Device IDs.
func (id *pluginID) NewNamespacedID(name string) string {
	return uuid.NewSHA1(id.uuid, []byte(name)).String()
}
