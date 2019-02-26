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
	"fmt"
	"github.com/vapor-ware/synse-server-grpc/go"
	"testing"

	"github.com/stretchr/testify/assert"
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
	}

	for i, c := range cases {
		tag, err := NewTag(c.tag)

		assert.Error(t, err, "case: %d", i)
		assert.Nil(t, tag, "case: %d", i)
	}
}

func TestNewTagFromGRPC(t *testing.T) {
	cases := []struct {
		tag *synse.V3Tag
		expected string
	}{
		{
			tag: &synse.V3Tag{},
			expected: "",
		},
		{
			tag: &synse.V3Tag{Label: "foo"},
			expected: "foo",
		},
		{
			tag: &synse.V3Tag{Namespace: "vapor", Label: "foo"},
			expected: "vapor/foo",
		},
		{
			tag: &synse.V3Tag{Annotation: "xyz", Label:"foo"},
			expected: "xyz:foo",
		},
		{
			tag: &synse.V3Tag{Namespace: "vapor", Annotation: "xyz", Label: "foo"},
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
		namespace string
		annotation string
		label string
	}{
		{
			label: "foo",
		},
		{
			annotation: "xyz",
			label: "foo",
		},
		{
			namespace: "vapor",
			label: "foo",
		},
		{
			namespace: "vapor",
			annotation: "xyz",
			label: "foo",
		},
	}

	for _, c := range cases {
		tag := Tag{
			Namespace: c.namespace,
			Annotation: c.annotation,
			Label: c.label,
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
		devices: []*Device{{id: "1"}, {id: "2"}},
	}

	set.Filter([]*Device{{id: "1"}, {id: "2"}, {id: "3"}, {id: "4"}})

	assert.Len(t, set.devices, 2)
}

func TestFilterSet_Filter2(t *testing.T) {
	set := filterSet{
		initialized: true,
		devices: []*Device{{id: "1"}, {id: "2"}, {id: "3"}, {id: "4"}},
	}

	set.Filter([]*Device{{id: "1"}, {id: "2"}})

	assert.Len(t, set.devices, 2)
}

func TestFilterSet_Filter3(t *testing.T) {
	set := filterSet{
		initialized: true,
		devices: []*Device{{id: "3"}, {id: "4"}},
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
			{Namespace: "foo", Annotation:"bar", Label:"baz"},
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
