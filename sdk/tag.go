// Synse SDK
// Copyright (c) 2017-2020 Vapor IO
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
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

// Tag component constants.
const (
	// Tag namespace constants
	TagNamespaceDefault = "default"
	TagNamespaceSystem  = "system"

	// Tag annotation constants
	TagAnnotationID   = "id"
	TagAnnotationType = "type"

	// Special tag components
	TagLabelAll = "**"
)

var (
	tagsTmpl = template.New("tags").Funcs(template.FuncMap{
		"env": os.Getenv,
	})
)

// Tag represents a group identifier which a Synse device can belong to.
type Tag struct {
	Namespace  string
	Annotation string
	Label      string

	string string
}

// NewTag creates a new Tag from a tag string.
func NewTag(tag string) (*Tag, error) {
	if tag == "" {
		return nil, fmt.Errorf("cannot create tag from empty string")
	}

	// First, attempt to parse the tag string as if it were a template - this could be
	// the case if part of the tag is templated out in config, e.g. "foo/bar:{{ env BAZ }}".
	tmpl, err := tagsTmpl.Parse(tag)
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, tag); err != nil {
		return nil, err
	}
	tag = buf.String()

	tag = strings.TrimSpace(tag)
	if strings.Contains(tag, " ") {
		log.WithField("tag", tag).Error("[tag] invalid: tag must not contain spaces")
		return nil, fmt.Errorf("tag must not contain spaces")
	}

	validTag := regexp.MustCompile(`(([^:/]+)/)?(([^:/]+):)?([^:/\s]+$)`)

	// The regular expression we match to has 5 groups:
	//   group 1: ((.+)/)?    -> namespace component with trailing slash
	//   group 2:  (.+)       -> namespace without trailing slash
	//   group 3: ((.+):)?    -> annotation component with trailing colon
	//   group 4:  (.+)       -> annotation without trailing colon
	//   group 5: ([^:/\s]+?) -> label
	//
	// We only care about group 2 (namespace), group 4 (annotation), and
	// group 5 (label), which are found in the corresponding indices.
	// (index 0 is the full match)
	matches := validTag.FindStringSubmatch(tag)

	// If we don't get the expected number of groups, the string does not
	// represent a tag we can do anything with.
	if len(matches) != 6 {
		log.WithField("tag", tag).Error("[tag] invalid: failed regex match")
		return nil, fmt.Errorf("invalid tag string (match check): %s", tag)
	}

	namespace := matches[2]
	annotation := matches[4]
	label := matches[5]

	// Make sure that the original tag does not have a namespace delimiter
	// if no namespace was matched. This is indicative of a malformed tag which
	// the regex may not have choked on.
	if strings.Contains(tag, "/") && namespace == "" {
		log.WithField("tag", tag).Error("[tag] invalid: failed namespace check")
		return nil, fmt.Errorf("invalid tag string (namespace check): %s", tag)
	}

	// Make sure that the original tag does not have an annotation delimiter
	// if no annotation was matched. This is indicative of a malformed tag which
	// the regex may not have choked on.
	if strings.Contains(tag, ":") && annotation == "" {
		log.WithField("tag", tag).Error("[tag] invalid: failed annotation check")
		return nil, fmt.Errorf("invalid tag string (annotation check): %s", tag)
	}

	// If no namespace is specified, use the default namespace.
	if namespace == "" {
		log.WithField("tag", tag).Debug("[tag] using default namespace for tag")
		namespace = TagNamespaceDefault
	}

	return &Tag{
		Namespace:  namespace,
		Annotation: annotation,
		Label:      label,
		string:     tag,
	}, nil
}

// NewTagFromGRPC creates a new Tag from the gRPC tag message.
func NewTagFromGRPC(tag *synse.V3Tag) *Tag {
	var tagString string
	if tag.Namespace != "" {
		tagString += tag.Namespace + "/"
	}
	if tag.Annotation != "" {
		tagString += tag.Annotation + ":"
	}
	tagString += tag.Label

	log.WithField("tag", tagString).Debug("[tag] created new tag from gRPC")
	return &Tag{
		Namespace:  tag.Namespace,
		Annotation: tag.Annotation,
		Label:      tag.Label,
		string:     tagString,
	}
}

// HasNamespace checks whether the tag has a namespace defined.
func (tag *Tag) HasNamespace() bool {
	return tag.Namespace != ""
}

// HasAnnotation checks whether the tag has an annotation defined.
func (tag *Tag) HasAnnotation() bool {
	return tag.Annotation != ""
}

// String prints the Tag in its string representation.
func (tag *Tag) String() string {
	return tag.string
}

// Encode translates the Tag to its corresponding gRPC message.
func (tag *Tag) Encode() *synse.V3Tag {
	return &synse.V3Tag{
		Namespace:  tag.Namespace,
		Annotation: tag.Annotation,
		Label:      tag.Label,
	}
}

// filterSet is a type which makes it easier to aggregate Devices based on tag filters.
//
// Tag filtering is subtractive, that is to say: get only the devices which match every
// specified tag. This essentially means that when joining Device slices for tag matches,
// we need to keep only those devices which are the same (set intersection).
type filterSet struct {
	devices     []*Device
	initialized bool
}

// Filter filters the filterSet's devices with the new device set provided. This is
// effectively a set intersection.
func (set *filterSet) Filter(devices []*Device) {
	// If the filterSet is not initialized, just update the internal device slice
	// with the provided devices.
	if len(set.devices) == 0 && !set.initialized {
		set.devices = devices
		set.initialized = true
		return
	}

	// We already have devices, so we need to get the intersection.
	var intersection []*Device
	for _, device := range set.devices {
		for _, d := range devices {
			if device.id == d.id {
				intersection = append(intersection, device)
				break
			}
		}
	}
	set.devices = intersection
}

// Results gets the filtered slice of Devices.
func (set *filterSet) Results() []*Device {
	return set.devices
}

// TagCache is a cache which can be used for looking up devices based on
// their tags.
type TagCache struct {
	// cache is the internal cache data structure which is used to do tag
	// routing to lookup matching devices.
	//
	// The outer map key is the namespace. The middle map key is the annotation.
	// The inner map key is the label.
	//
	// With this cache structure, we can find all devices which match a tag
	// by decomposing the tag into its searchable components and traversing
	// the cache.
	cache map[string]map[string]map[string][]*Device
}

// NewTagCache creates a new TagCache instance.
func NewTagCache() *TagCache {
	return &TagCache{
		cache: make(map[string]map[string]map[string][]*Device),
	}
}

// Add adds a device to the tag cache for the specified tag.
func (cache *TagCache) Add(tag *Tag, device *Device) {
	if tag.Label == TagLabelAll {
		log.WithFields(log.Fields{
			"tag":    tag.String(),
			"device": device.GetID(),
		}).Debug("[tag] will not cache device for 'all' label")
		return
	}

	cacheLog := log.WithFields(log.Fields{
		"namespace":  tag.Namespace,
		"annotation": tag.Annotation,
		"label":      tag.Label,
		"device":     device.id,
	})

	annotations, exists := cache.cache[tag.Namespace]
	if !exists {
		// If the namespace doesn't exist, add it with the rest of the tag info.
		cache.cache[tag.Namespace] = map[string]map[string][]*Device{
			tag.Annotation: {tag.Label: {device}},
		}
		cacheLog.Debug("[tag] added new namespace to tag cache")
		return
	}

	labels, exists := annotations[tag.Annotation]
	if !exists {
		// If the annotation doesn't exist, add it with the rest of the tag info.
		annotations[tag.Annotation] = map[string][]*Device{
			tag.Label: {device},
		}
		cacheLog.Debug("[tag] added new annotation to tag cache")
		return
	}

	devices, exists := labels[tag.Label]
	if !exists {
		// If the label doesn't exist, add it with the device.
		labels[tag.Label] = []*Device{device}
		cacheLog.Debug("[tag] added new label to tag cache")
		return
	}

	// If we get here, the namespace, annotation, and label all exist. We just want
	// to add the device, but only if it is not already there.
	var duplicate bool
	for _, d := range devices {
		if d.id == device.id {
			duplicate = true
			cacheLog.Debug("[tag] device already exists in cache, skipping")
			break
		}
	}
	if !duplicate {
		labels[tag.Label] = append(devices, device)
		cacheLog.Debug("[tag] added device existing label in tag cache")
	}
}

// GetDevicesFromStrings gets the list of Devices which match the given set
// of tag strings.
func (cache *TagCache) GetDevicesFromStrings(tags ...string) ([]*Device, error) {
	var t = make([]*Tag, len(tags))
	for i, tag := range tags {
		nt, err := NewTag(tag)
		if err != nil {
			return nil, err
		}
		t[i] = nt
	}
	return cache.GetDevicesFromTags(t...), nil
}

// GetDevicesFromTags gets the list of Devices which match the given set of tags.
func (cache *TagCache) GetDevicesFromTags(tags ...*Tag) []*Device {
	var deviceSet filterSet

	for _, tag := range tags {
		annotations, exists := cache.cache[tag.Namespace]
		if !exists {
			// If a tag namespace is specified which doesn't exist in the cache,
			// there is no way a Device can match to it, so stop searching.
			return nil
		}

		// If the label specifies all devices, we will want to get all the devices
		// in the namespace if no annotation is defined or all devices in the the
		// namespace with the given annotation if the annotation is defined.
		if tag.Label == TagLabelAll {
			var devices []*Device

			if tag.HasAnnotation() {
				labels, exists := annotations[tag.Annotation]
				if !exists {
					// If a tag annotation is specified which doesn't exist in the cache,
					// there is no way a Device can match to it, so stop searching.
					return nil
				}
				for _, x := range labels {
					devices = append(devices, x...)
				}
				deviceSet.Filter(devices)
				continue
			}

			devices = cache.GetDevicesFromNamespace(tag.Namespace)
			deviceSet.Filter(devices)
			continue
		}

		labels, exists := annotations[tag.Annotation]
		if !exists {
			// If a tag annotation is specified which doesn't exist in the cache,
			// there is no way a Device can match to it, so stop searching.
			return nil
		}

		devices, exists := labels[tag.Label]
		if !exists {
			// If a tag label is specified which doesn't exist in the cache,
			// there is no way a Device can match to it, so stop searching.
			return nil
		}

		deviceSet.Filter(devices)
	}

	return deviceSet.Results()
}

// GetDevicesFromNamespace gets the devices for the specified namespaces.
func (cache *TagCache) GetDevicesFromNamespace(namespaces ...string) []*Device {
	// Initially, store the devices in a map. This will allow us to remove duplicates
	// which may be present, as a device can have a tag in multiple namespaces.
	var deviceMap = make(map[string]*Device)

	for _, ns := range namespaces {
		annotations, exists := cache.cache[ns]
		if !exists {
			continue
		}

		for _, a := range annotations {
			for _, l := range a {
				for _, device := range l {
					deviceMap[device.GetID()] = device
				}
			}
		}
	}

	// Make the final slice of devices to return.
	var devices = make([]*Device, 0, len(deviceMap))
	for _, d := range deviceMap {
		devices = append(devices, d)
	}
	return devices
}

// DeviceSelectorToTags is a utility that converts a gRPC device selector message
// into its corresponding tags.
func DeviceSelectorToTags(selector *synse.V3DeviceSelector) []*Tag {
	tag := DeviceSelectorToID(selector)
	if tag != nil {
		return []*Tag{tag}
	}

	var tags = make([]*Tag, len(selector.Tags))
	for i, t := range selector.Tags {
		tags[i] = NewTagFromGRPC(t)
	}
	return tags
}

// DeviceSelectorToID is a utility which converts a gRPC device selector message
// into a corresponding ID tag. If the selector has no value in the Id field, this
// will return nil.
func DeviceSelectorToID(selector *synse.V3DeviceSelector) *Tag {
	if selector.Id != "" {
		if len(selector.Tags) > 0 {
			log.WithFields(log.Fields{
				"id":   selector.Id,
				"tags": selector.Tags,
			}).Warn("[tags] device selector specifies id and tags; only using id (tags ignored)")
		}
		return newIDTag(selector.Id)
	}
	return nil
}

// newIDTag creates a new Tag for a device ID. These tags are auto-generated
// by the SDK and are considered system-wide tags.
func newIDTag(deviceID string) *Tag {
	return &Tag{
		Namespace:  TagNamespaceSystem,
		Annotation: TagAnnotationID,
		Label:      deviceID,
		string:     fmt.Sprintf("%s/%s:%s", TagNamespaceSystem, TagAnnotationID, deviceID),
	}
}

// newTypeTag creates a new Tag for a device Type. These tags are auto-generated
// by the SDK and are considered system-wide tags.
func newTypeTag(deviceType string) *Tag {
	return &Tag{
		Namespace:  TagNamespaceSystem,
		Annotation: TagAnnotationType,
		Label:      deviceType,
		string:     fmt.Sprintf("%s/%s:%s", TagNamespaceSystem, TagAnnotationType, deviceType),
	}
}
