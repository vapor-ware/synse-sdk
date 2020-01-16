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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactPasswords(t *testing.T) {

	var nilMap map[string]interface{}
	var nilSlice []interface{}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: nil,
		},
		{
			name:     "boolean",
			input:    true,
			expected: true,
		},
		{
			name:     "string",
			input:    "test-string",
			expected: "test-string",
		},
		{
			name:     "map with no password",
			input:    map[string]interface{}{"key": "value"},
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "map with key pass, value string",
			input:    map[string]interface{}{"key": "value", "pass": "foobar"},
			expected: map[string]interface{}{"key": "value", "pass": "REDACTED"},
		},
		{
			name:     "map with key PASS, value int",
			input:    map[string]interface{}{"key": "value", "PASS": 123},
			expected: map[string]interface{}{"key": "value", "PASS": "REDACTED"},
		},
		{
			name:     "map with key Password, value string",
			input:    map[string]interface{}{"key": "value", "Password": "password"},
			expected: map[string]interface{}{"key": "value", "Password": "REDACTED"},
		},
		{
			name:     "map with key User Password, value map",
			input:    map[string]interface{}{"key": "value", "User Password": map[string]interface{}{"foo": "bar"}},
			expected: map[string]interface{}{"key": "value", "User Password": "REDACTED"},
		},
		{
			name:     "map with key userpass, value slice",
			input:    map[string]interface{}{"key": "value", "userpass": []interface{}{1, 2, 3}},
			expected: map[string]interface{}{"key": "value", "userpass": "REDACTED"},
		},
		{
			name:     "map with nested map with password",
			input:    map[string]interface{}{"key": map[string]interface{}{"key": "value", "pass": "foo"}},
			expected: map[string]interface{}{"key": map[string]interface{}{"key": "value", "pass": "REDACTED"}},
		},
		{
			name:     "map with nested map with no password",
			input:    map[string]interface{}{"key": map[string]interface{}{"key": "value", "bar": "foo"}},
			expected: map[string]interface{}{"key": map[string]interface{}{"key": "value", "bar": "foo"}},
		},
		{
			name:     "map with nested slice with nested map with password",
			input:    map[string]interface{}{"foo": []interface{}{map[string]interface{}{"pass": "foo", "other": "bar"}}},
			expected: map[string]interface{}{"foo": []interface{}{map[string]interface{}{"pass": "REDACTED", "other": "bar"}}},
		},
		{
			name:     "map with empty slice",
			input:    map[string]interface{}{"key": []interface{}{}},
			expected: map[string]interface{}{"key": []interface{}{}},
		},
		{
			name:     "map with empty map",
			input:    map[string]interface{}{"key": map[string]interface{}{}},
			expected: map[string]interface{}{"key": map[string]interface{}{}},
		},
		{
			name:     "map with nil slice",
			input:    map[string]interface{}{"key": nilSlice},
			expected: map[string]interface{}{"key": nilSlice},
		},
		{
			name:     "map with nil map",
			input:    map[string]interface{}{"key": nilMap},
			expected: map[string]interface{}{"key": nilMap},
		},
		{
			name:     "slice with no maps",
			input:    []interface{}{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "slice with map, no password",
			input:    []interface{}{map[string]interface{}{"foo": "bar", "abc": "123"}},
			expected: []interface{}{map[string]interface{}{"foo": "bar", "abc": "123"}},
		},
		{
			name:     "slice with map, key pass, value string",
			input:    []interface{}{map[string]interface{}{"foo": "bar", "pass": "123"}},
			expected: []interface{}{map[string]interface{}{"foo": "bar", "pass": "REDACTED"}},
		},
		{
			name:     "slice with nested slice, no pass",
			input:    []interface{}{[]interface{}{"a", "b", "c"}},
			expected: []interface{}{[]interface{}{"a", "b", "c"}},
		},
		{
			name:     "slice with nested slice, with password",
			input:    []interface{}{[]interface{}{map[string]interface{}{"pass": "foo"}}},
			expected: []interface{}{[]interface{}{map[string]interface{}{"pass": "REDACTED"}}},
		},
		{
			name:     "slice with empty map",
			input:    []interface{}{map[string]interface{}{}},
			expected: []interface{}{map[string]interface{}{}},
		},
		{
			name:     "slice with empty slice",
			input:    []interface{}{[]interface{}{}},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "slice with nil map",
			input:    []interface{}{nilMap},
			expected: []interface{}{nilMap},
		},
		{
			name:     "slice with nil slice",
			input:    []interface{}{nilSlice},
			expected: []interface{}{nilSlice},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			redacted := RedactPasswords(test.input)
			assert.Equal(t, test.expected, redacted)
		})
	}
}
