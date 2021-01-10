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
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

func TestNewTag(t *testing.T) {
	cases := []struct {
		tag        string
		namespace  string
		annotation string
		label      string
		str        string
	}{
		{
			tag:        "foo",
			namespace:  "default",
			annotation: "",
			label:      "foo",
		},
		{
			tag:        "bar",
			namespace:  "default",
			annotation: "",
			label:      "bar",
		},
		{
			tag:        "a/foo",
			namespace:  "a",
			annotation: "",
			label:      "foo",
		},
		{
			tag:        "b/bar",
			namespace:  "b",
			annotation: "",
			label:      "bar",
		},
		{
			tag:        "x:foo",
			namespace:  "default",
			annotation: "x",
			label:      "foo",
		},
		{
			tag:        "y:bar",
			namespace:  "default",
			annotation: "y",
			label:      "bar",
		},
		{
			tag:        "a/x:foo",
			namespace:  "a",
			annotation: "x",
			label:      "foo",
		},
		{
			tag:        "b/y:bar",
			namespace:  "b",
			annotation: "y",
			label:      "bar",
		},
		{
			tag:        "a-b/x-y:m-n",
			namespace:  "a-b",
			annotation: "x-y",
			label:      "m-n",
		}, {
			tag:        "a.b/x.y:m.n",
			namespace:  "a.b",
			annotation: "x.y",
			label:      "m.n",
		},
		{
			tag:        "  yankee/hotel:foxtrot  ",
			namespace:  "yankee",
			annotation: "hotel",
			label:      "foxtrot",
			str:        "yankee/hotel:foxtrot",
		},
	}

	for i, c := range cases {
		tag, err := NewTag(c.tag)

		assert.NoError(t, err, "case: %d", i)
		assert.Equal(t, c.namespace, tag.Namespace, "case: %d", i)
		assert.Equal(t, c.annotation, tag.Annotation, "case: %d", i)
		assert.Equal(t, c.label, tag.Label, "case: %d", i)

		str := c.str
		if c.str == "" {
			str = c.tag
		}

		assert.Equal(t, str, tag.String(), "case: %d", i)
	}
}

func TestNewTag_Error(t *testing.T) {
	cases := []struct {
		tag string
	}{
		{tag: ""},
		{tag: "a//b"},
		{tag: "a::b"},
		{tag: "a/b:"},
		{tag: "/"},
		{tag: "//"},
		{tag: ":"},
		{tag: "::"},
		{tag: "vaporio/contains spaces:foo"},
	}

	for i, c := range cases {
		tag, err := NewTag(c.tag)

		assert.Error(t, err, "case: %d", i)
		assert.Nil(t, tag, "case: %d", i)
	}
}

func TestNewTagWithEnv(t *testing.T) {
	cases := []struct {
		tag        string
		namespace  string
		annotation string
		label      string
		str        string
	}{
		{
			tag:        `{{ env "FOO" }}`,
			namespace:  "default",
			annotation: "",
			label:      "foo",
			str:        "foo",
		},
		{
			tag:        `foo/{{ env "BAR" }}`,
			namespace:  "foo",
			annotation: "",
			label:      "bar",
			str:        "foo/bar",
		},
		{
			tag:        `{{env "FOO"}}/{{env "BAR"}}`,
			namespace:  "foo",
			annotation: "",
			label:      "bar",
			str:        "foo/bar",
		},
		{
			tag:        ` testing/{{ env "FOO"}}:{{ env "TEST_ENV_VAL_1" }} `,
			namespace:  "testing",
			annotation: "foo",
			label:      "1",
			str:        "testing/foo:1",
		},
		{
			tag:        `{{env "FOO"}}/{{env "BAR"}}:{{env "TEST_ENV_VAL_2"}}`,
			namespace:  "foo",
			annotation: "bar",
			label:      "2",
			str:        "foo/bar:2",
		},
	}

	testEnv := map[string]string{
		"FOO":            "foo",
		"BAR":            "bar",
		"TEST_ENV_VAL_1": "1",
		"TEST_ENV_VAL_2": "2",
	}
	// Setup the environment for the test case.
	for k, v := range testEnv {
		err := os.Setenv(k, v)
		assert.NoError(t, err)
	}
	defer func() {
		for k := range testEnv {
			err := os.Unsetenv(k)
			assert.NoError(t, err)
		}
	}()

	for i, c := range cases {
		tag, err := NewTag(c.tag)

		assert.NoError(t, err, "case: %d", i)
		assert.Equal(t, c.namespace, tag.Namespace, "case: %d", i)
		assert.Equal(t, c.annotation, tag.Annotation, "case: %d", i)
		assert.Equal(t, c.label, tag.Label, "case: %d", i)

		str := c.str
		if c.str == "" {
			str = c.tag
		}

		assert.Equal(t, str, tag.String(), "case: %d", i)
	}
}

func TestNewTagWithEnv_Error(t *testing.T) {
	cases := []struct {
		tag string
	}{
		{tag: `default/{{ env "VALUE_NOT_SET" }}`},
		{tag: `{{ env "VALUE_NOT_SET" }}`},
		{tag: `{{ env "VALUE_NOT_SET" }}/foo`},
	}

	for i, c := range cases {
		tag, err := NewTag(c.tag)

		assert.Error(t, err, "case: %d", i)
		assert.Nil(t, tag, "case: %d", i)
	}
}

func TestNewTagFromGRPC(t *testing.T) {
	cases := []struct {
		tag      *synse.V3Tag
		expected string
	}{
		{
			tag:      &synse.V3Tag{},
			expected: "",
		},
		{
			tag:      &synse.V3Tag{Label: "foo"},
			expected: "foo",
		},
		{
			tag:      &synse.V3Tag{Namespace: "vapor", Label: "foo"},
			expected: "vapor/foo",
		},
		{
			tag:      &synse.V3Tag{Annotation: "xyz", Label: "foo"},
			expected: "xyz:foo",
		},
		{
			tag:      &synse.V3Tag{Namespace: "vapor", Annotation: "xyz", Label: "foo"},
			expected: "vapor/xyz:foo",
		},
	}

	for _, c := range cases {
		tag := NewTagFromGRPC(c.tag)
		assert.Equal(t, c.expected, tag.String())
	}
}

func TestTag_HasAnnotation(t *testing.T) {
	tag := Tag{Annotation: "foo"}
	assert.True(t, tag.HasAnnotation())

	tag = Tag{}
	assert.False(t, tag.HasAnnotation())
}

func TestTag_HasNamespace(t *testing.T) {
	tag := Tag{Namespace: "foo"}
	assert.True(t, tag.HasNamespace())

	tag = Tag{}
	assert.False(t, tag.HasNamespace())
}

func TestTag_String(t *testing.T) {
	tag := Tag{string: "foo/bar"}
	assert.Equal(t, "foo/bar", tag.String())

	tag = Tag{}
	assert.Equal(t, "", tag.String())
}

func TestTag_Encode(t *testing.T) {
	cases := []struct {
		namespace  string
		annotation string
		label      string
	}{
		{
			label: "foo",
		},
		{
			annotation: "xyz",
			label:      "foo",
		},
		{
			namespace: "vapor",
			label:     "foo",
		},
		{
			namespace:  "vapor",
			annotation: "xyz",
			label:      "foo",
		},
	}

	for _, c := range cases {
		tag := Tag{
			Namespace:  c.namespace,
			Annotation: c.annotation,
			Label:      c.label,
		}

		encoded := tag.Encode()
		assert.Equal(t, c.namespace, encoded.Namespace)
		assert.Equal(t, c.annotation, encoded.Annotation)
		assert.Equal(t, c.label, encoded.Label)
	}
}

func TestFilterSet_Filter_firstFilter(t *testing.T) {
	set := filterSet{}

	assert.Empty(t, set.devices)
	assert.False(t, set.initialized)

	set.Filter([]*Device{{id: "1"}, {id: "2"}})

	assert.Len(t, set.devices, 2)
	assert.True(t, set.initialized)
}

func TestFilterSet_Filter1(t *testing.T) {
	set := filterSet{
		initialized: true,
		devices:     []*Device{{id: "1"}, {id: "2"}},
	}

	set.Filter([]*Device{{id: "1"}, {id: "2"}, {id: "3"}, {id: "4"}})

	assert.Len(t, set.devices, 2)
}

func TestFilterSet_Filter2(t *testing.T) {
	set := filterSet{
		initialized: true,
		devices:     []*Device{{id: "1"}, {id: "2"}, {id: "3"}, {id: "4"}},
	}

	set.Filter([]*Device{{id: "1"}, {id: "2"}})

	assert.Len(t, set.devices, 2)
}

func TestFilterSet_Filter3(t *testing.T) {
	set := filterSet{
		initialized: true,
		devices:     []*Device{{id: "3"}, {id: "4"}},
	}

	set.Filter([]*Device{{id: "1"}, {id: "2"}})

	assert.Len(t, set.devices, 0)
}

func TestFilterSet_Results(t *testing.T) {
	set := filterSet{
		devices: []*Device{{id: "1"}, {id: "2"}},
	}

	results := set.Results()
	assert.Len(t, results, 2)
}

func TestFilterSet_Results2(t *testing.T) {
	set := filterSet{}

	results := set.Results()
	assert.Len(t, results, 0)
}

func TestNewTagCache(t *testing.T) {
	cache := NewTagCache()
	assert.Empty(t, cache.cache)
}

func TestTagCache_Add_labelAll(t *testing.T) {
	// Test that nothing is added to a cache if the tag uses the special "all" label
	cache := &TagCache{}

	device := Device{id: "1234"}
	tag := Tag{Namespace: "foo", Annotation: "bar", Label: TagLabelAll}

	assert.Empty(t, cache.cache)
	cache.Add(&tag, &device)
	assert.Empty(t, cache.cache)
}

func TestTagCache_Add_newNamespace(t *testing.T) {
	// Test adding a device with a namespace that is not yet in the cache.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{},
	}

	device := Device{id: "1234"}
	tag := Tag{Namespace: "foo", Annotation: "bar", Label: "baz"}

	assert.Empty(t, cache.cache)
	cache.Add(&tag, &device)

	assert.Len(t, cache.cache, 1)
	assert.Contains(t, cache.cache, "foo")
	assert.Contains(t, cache.cache["foo"], "bar")
	assert.Contains(t, cache.cache["foo"]["bar"], "baz")
	assert.Len(t, cache.cache["foo"]["bar"]["baz"], 1)
}

func TestTagCache_Add_newAnnotation(t *testing.T) {
	// Test adding a device with an annotation that is not yet in the cache.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {},
		},
	}

	device := Device{id: "1234"}
	tag := Tag{Namespace: "foo", Annotation: "bar", Label: "baz"}

	cache.Add(&tag, &device)

	assert.Len(t, cache.cache, 1)
	assert.Contains(t, cache.cache, "foo")
	assert.Contains(t, cache.cache["foo"], "bar")
	assert.Contains(t, cache.cache["foo"]["bar"], "baz")
	assert.Len(t, cache.cache["foo"]["bar"]["baz"], 1)
}

func TestTagCache_Add_newLabel(t *testing.T) {
	// Test adding a new device when the label is not yet in the cache.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {},
			},
		},
	}

	device := Device{id: "1234"}
	tag := Tag{Namespace: "foo", Annotation: "bar", Label: "baz"}

	cache.Add(&tag, &device)

	assert.Len(t, cache.cache, 1)
	assert.Contains(t, cache.cache, "foo")
	assert.Contains(t, cache.cache["foo"], "bar")
	assert.Contains(t, cache.cache["foo"]["bar"], "baz")
	assert.Len(t, cache.cache["foo"]["bar"]["baz"], 1)
}

func TestTagCache_Add_newTag2(t *testing.T) {
	// Test adding a new device when the label already exists in the cache.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
					},
				},
			},
		},
	}

	device := Device{id: "1234"}
	tag := Tag{Namespace: "foo", Annotation: "bar", Label: "baz"}

	cache.Add(&tag, &device)

	assert.Len(t, cache.cache, 1)
	assert.Contains(t, cache.cache, "foo")
	assert.Contains(t, cache.cache["foo"], "bar")
	assert.Contains(t, cache.cache["foo"]["bar"], "baz")
	assert.Len(t, cache.cache["foo"]["bar"]["baz"], 2)
}

func TestTagCache_Add_duplicate(t *testing.T) {
	// Test adding a device when that device is already added for the tag.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
					},
				},
			},
		},
	}

	device := Device{id: "xyz"}
	tag := Tag{Namespace: "foo", Annotation: "bar", Label: "baz"}

	cache.Add(&tag, &device)

	assert.Len(t, cache.cache, 1)
	assert.Contains(t, cache.cache, "foo")
	assert.Contains(t, cache.cache["foo"], "bar")
	assert.Contains(t, cache.cache["foo"]["bar"], "baz")
	assert.Len(t, cache.cache["foo"]["bar"]["baz"], 1)
}

func TestTagCache_GetDevicesFromTags_noTags(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
					},
				},
			},
		},
	}

	devices := cache.GetDevicesFromTags()
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromTags_noNsMatch(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
					},
				},
			},
		},
	}

	tag := Tag{Namespace: "a", Annotation: "bar", Label: "baz"}
	devices := cache.GetDevicesFromTags(&tag)
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromTags_noAnnotationMatch(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
					},
				},
			},
		},
	}

	tag := Tag{Namespace: "foo", Annotation: "b", Label: "baz"}
	devices := cache.GetDevicesFromTags(&tag)
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromTags_noLabelMatch(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
					},
				},
			},
		},
	}

	tag := Tag{Namespace: "foo", Annotation: "bar", Label: "c"}
	devices := cache.GetDevicesFromTags(&tag)
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromTags_labelAll_noNsMatch(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
						&Device{id: "abc"},
					},
				},
			},
		},
	}

	tag := Tag{Namespace: "a", Annotation: "bar", Label: "**"}
	devices := cache.GetDevicesFromTags(&tag)
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromTags_labelAll_noAnnotationMatch(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
						&Device{id: "abc"},
					},
				},
			},
		},
	}

	tag := Tag{Namespace: "foo", Annotation: "b", Label: "**"}
	devices := cache.GetDevicesFromTags(&tag)
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromTags(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
						&Device{id: "abc"},
					},
				},
				"baz": {
					"bat": {
						&Device{id: "456"},
						&Device{id: "abc"},
					},
					"b0t": {
						&Device{id: "456"},
						&Device{id: "xyz"},
					},
				},
			},
			"default": {
				"": {
					"vapor": {
						&Device{id: "xyz"},
						&Device{id: "123"},
					},
				},
				"type": {
					"led": {
						&Device{id: "xyz"},
						&Device{id: "123"},
					},
				},
			},
		},
	}

	tag1 := Tag{Namespace: "foo", Annotation: "bar", Label: "baz"}
	tag2 := Tag{Namespace: "default", Annotation: "", Label: "vapor"}
	devices := cache.GetDevicesFromTags(&tag1, &tag2)
	assert.Len(t, devices, 1)
	assert.Equal(t, "xyz", devices[0].id)
}

func TestTagCache_GetDevicesFromTags_all(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
						&Device{id: "abc"},
					},
				},
				"baz": {
					"bat": {
						&Device{id: "456"},
						&Device{id: "abc"},
					},
					"b0t": {
						&Device{id: "456"},
						&Device{id: "xyz"},
					},
				},
			},
			"default": {
				"": {
					"vapor": {
						&Device{id: "xyz"},
					},
				},
				"type": {
					"led": {
						&Device{id: "xyz"},
						&Device{id: "123"},
					},
				},
			},
		},
	}

	tag := Tag{Namespace: "foo", Annotation: "bar", Label: "**"}
	devices := cache.GetDevicesFromTags(&tag)
	assert.Len(t, devices, 2)

	tag = Tag{Namespace: "foo", Label: "**"}
	devices = cache.GetDevicesFromTags(&tag)
	assert.Len(t, devices, 3)

	tag1 := Tag{Namespace: "foo", Label: "**"}
	tag2 := Tag{Namespace: "default", Annotation: "type", Label: "**"}
	devices = cache.GetDevicesFromTags(&tag1, &tag2)
	assert.Len(t, devices, 1)
}

func TestTagCache_GetDevicesFromStrings_noTags(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
						&Device{id: "abc"},
					},
				},
			},
		},
	}

	devices, err := cache.GetDevicesFromStrings()
	assert.NoError(t, err)
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromStrings_invalidTag(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
						&Device{id: "abc"},
					},
				},
			},
		},
	}

	devices, err := cache.GetDevicesFromStrings("not a tag string")
	assert.Error(t, err)
	assert.Nil(t, devices)
}

func TestTagCache_GetDevicesFromStrings_validTag(t *testing.T) {
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
						&Device{id: "abc"},
					},
				},
			},
		},
	}

	devices, err := cache.GetDevicesFromStrings("foo/bar:baz")
	assert.NoError(t, err)
	assert.Len(t, devices, 2)
}

func TestTagCache_GetDevicesFromNamespace_empty(t *testing.T) {
	// No namespaces defined in the cache.
	cache := &TagCache{}

	devices := cache.GetDevicesFromNamespace("foo", "bar")
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromNamespace_noNsSpecified(t *testing.T) {
	// No namespaces defined in the cache.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
					},
				},
			},
		},
	}

	devices := cache.GetDevicesFromNamespace()
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromNamespace_nsNotExist(t *testing.T) {
	// The specified namespace does not exist in the cache.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"baz": {
						&Device{id: "xyz"},
					},
				},
			},
		},
	}

	devices := cache.GetDevicesFromNamespace("abc")
	assert.Empty(t, devices)
}

func TestTagCache_GetDevicesFromNamespace_multipleDevices(t *testing.T) {
	// The namespace contains multiple devices.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"a": {
						&Device{id: "xyz"},
					},
					"b": {
						&Device{id: "123"},
					},
				},
				"baz": {
					"c": {
						&Device{id: "abc"},
					},
				},
			},
		},
	}

	devices := cache.GetDevicesFromNamespace("foo")
	assert.Len(t, devices, 3)
}

func TestTagCache_GetDevicesFromNamespace_multipleNamespaces(t *testing.T) {
	// Multiple namespaces are specified, each with their own devices.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"a": {
						&Device{id: "xyz"},
					},
					"b": {
						&Device{id: "123"},
					},
				},
				"baz": {
					"c": {
						&Device{id: "abc"},
					},
				},
			},
			"default": {
				"": {
					"x": {
						&Device{id: "456"},
					},
					"y": {
						&Device{id: "789"},
					},
				},
				"vapor": {
					"z": {
						&Device{id: "vapor1"},
					},
				},
			},
		},
	}

	devices := cache.GetDevicesFromNamespace("foo")
	assert.Len(t, devices, 3)

	devices = cache.GetDevicesFromNamespace("default")
	assert.Len(t, devices, 3)

	devices = cache.GetDevicesFromNamespace("foo", "default")
	assert.Len(t, devices, 6)
}

func TestTagCache_GetDevicesFromNamespace_multipleNamespacesDuplicate(t *testing.T) {
	// Multiple namespaces are specified, with some device overlap.
	cache := &TagCache{
		cache: map[string]map[string]map[string][]*Device{
			"foo": {
				"bar": {
					"a": {
						&Device{id: "xyz"},
					},
					"b": {
						&Device{id: "123"},
						&Device{id: "456"},
						&Device{id: "789"},
					},
				},
				"baz": {
					"c": {
						&Device{id: "abc"},
					},
				},
			},
			"default": {
				"": {
					"x": {
						&Device{id: "456"},
					},
					"y": {
						&Device{id: "789"},
					},
				},
				"vapor": {
					"z": {
						&Device{id: "vapor1"},
						&Device{id: "abc"},
					},
				},
			},
		},
	}

	devices := cache.GetDevicesFromNamespace("foo")
	assert.Len(t, devices, 5)

	devices = cache.GetDevicesFromNamespace("default")
	assert.Len(t, devices, 4)

	devices = cache.GetDevicesFromNamespace("foo", "default")
	assert.Len(t, devices, 6)
}

func TestDeviceSelectorToTags_withID(t *testing.T) {
	tags := DeviceSelectorToTags(&synse.V3DeviceSelector{
		Id: "1234",
	})

	assert.Len(t, tags, 1)

	tag := tags[0]
	assert.Equal(t, TagNamespaceSystem, tag.Namespace)
	assert.Equal(t, TagAnnotationID, tag.Annotation)
	assert.Equal(t, "1234", tag.Label)
	assert.Equal(t, "system/id:1234", tag.String())
}

func TestDeviceSelectorToTags(t *testing.T) {
	tags := DeviceSelectorToTags(&synse.V3DeviceSelector{
		Tags: []*synse.V3Tag{
			{Namespace: "vapor", Label: "0"},
			{Namespace: "vapor", Label: "1"},
			{Namespace: "vapor", Label: "2"},
		},
	})

	assert.Len(t, tags, 3)
	for i, tag := range tags {
		assert.Equal(t, "vapor", tag.Namespace)
		assert.Equal(t, "", tag.Annotation)
		assert.Equal(t, fmt.Sprintf("%d", i), tag.Label)
	}
}

func TestDeviceSelectorToID_noID(t *testing.T) {
	tag := DeviceSelectorToID(&synse.V3DeviceSelector{})
	assert.Nil(t, tag)
}

func TestDeviceSelectorToID(t *testing.T) {
	tag := DeviceSelectorToID(&synse.V3DeviceSelector{
		Id: "1234",
	})

	assert.Equal(t, TagNamespaceSystem, tag.Namespace)
	assert.Equal(t, TagAnnotationID, tag.Annotation)
	assert.Equal(t, "1234", tag.Label)
	assert.Equal(t, "system/id:1234", tag.String())
}

func TestDeviceSelectorToID_withTags(t *testing.T) {
	tag := DeviceSelectorToID(&synse.V3DeviceSelector{
		Id: "1234",
		Tags: []*synse.V3Tag{
			{Namespace: "foo", Annotation: "bar", Label: "baz"},
		},
	})

	assert.Equal(t, TagNamespaceSystem, tag.Namespace)
	assert.Equal(t, TagAnnotationID, tag.Annotation)
	assert.Equal(t, "1234", tag.Label)
	assert.Equal(t, "system/id:1234", tag.String())
}

func TestNewIDTag(t *testing.T) {
	tag := newIDTag("1234")
	assert.Equal(t, TagNamespaceSystem, tag.Namespace)
	assert.Equal(t, TagAnnotationID, tag.Annotation)
	assert.Equal(t, "1234", tag.Label)
	assert.Equal(t, "system/id:1234", tag.String())
}

func TestNewTypeTag(t *testing.T) {
	tag := newTypeTag("foo")
	assert.Equal(t, TagNamespaceSystem, tag.Namespace)
	assert.Equal(t, TagAnnotationType, tag.Annotation)
	assert.Equal(t, "foo", tag.Label)
	assert.Equal(t, "system/type:foo", tag.String())
}

func TestConcurrentMapWrites(t *testing.T) {
	testCases := []struct {
		tag      string
		expected *Tag
	}{
		{
			"a",
			&Tag{
				Namespace: "default",
				Label:     "a",
			},
		},
		{
			"a/b",
			&Tag{
				Namespace: "a",
				Label:     "b",
			},
		},
		{
			"a/b:c",
			&Tag{
				Namespace:  "a",
				Annotation: "b",
				Label:      "c",
			},
		},
		{
			`a/b:{{ identity "foobar" }}`,
			&Tag{
				Namespace:  "a",
				Annotation: "b",
				Label:      "foobar",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.tag, func(t *testing.T) {
			// Run tests in parallel so we hit the case where template.Parse/template.Execute
			// would be called concurrently.
			t.Parallel()

			tag, err := NewTag(tc.tag)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.Namespace, tag.Namespace)
			assert.Equal(t, tc.expected.Annotation, tag.Annotation)
			assert.Equal(t, tc.expected.Label, tag.Label)

		})
	}
}

//
// Benchmarks
//

// For context of how these benchmarks are written, see:
// https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go

var benchmarkTag *Tag

func BenchmarkNewTag(b *testing.B) {
	var t *Tag
	for n := 0; n < b.N; n++ {
		// Always record the result to prevent the compiler from eliminating
		// the function call.
		t, _ = NewTag("abc/def:ghi")
	}

	// Always store the result to a package-level var so the compiler doesn't
	// eliminate the Benchmark itself.
	benchmarkTag = t
}
