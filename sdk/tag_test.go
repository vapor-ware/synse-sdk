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
