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

package utils

import (
	"reflect"
	"strings"
)

// RedactPasswords redacts any map fields where the key contains the substring
// "pass" (case-insensitive). It traverses through any slice or map within to
// search for fields to redact.
//
// This does not make any attempt to find other potential passwords as
// magic strings. via regex, or via entropy. This is just meant to cover the
// basic case of "pass": "foo" within various config locations where it is
// likely to exist and should not be leaked out into logs.
//
// This function is likely very inefficient due to the interface casting and
// the need to copy the lists/maps provided so that it does not overwrite the
// original list/map.
func RedactPasswords(m interface{}) interface{} {

	switch m.(type) {
	case map[string]interface{}:
		redacted := map[string]interface{}{}
		for k, v := range m.(map[string]interface{}) {
			redacted[k] = v
		}
		traverseMap(redacted)
		return redacted

	case []interface{}:
		var redacted []interface{}
		for _, v := range m.([]interface{}) {
			redacted = append(redacted, RedactPasswords(v))
		}
		traverseSlice(redacted)
		return redacted

	case []map[string]interface{}:
		var redacted []interface{}
		for _, i := range m.([]map[string]interface{}) {
			mapCopy := map[string]interface{}{}
			for k, v := range i {
				mapCopy[k] = v
			}
			redacted = append(redacted, mapCopy)
		}
		traverseSlice(redacted)
		return redacted

	default:
		return m
	}
}

// traverseMap iterates through all keys and values in a map[string]interface{},
// replacing passwords with a redacted string. If it finds a nested
// map[string]interface{} we recurse into it.
func traverseMap(m map[string]interface{}) {
	for k, v := range m {

		// If the key contains the string "pass" (case-insensitive), we substitute
		// with the string REDACTED
		if strings.Contains(strings.ToLower(k), "pass") {
			// Redact the data whatever it is.
			m[k] = "REDACTED"
			continue
		}

		// Is this a map of [string]interface{}?
		vvalue := reflect.ValueOf(v)
		vkind := vvalue.Kind()
		if vkind == reflect.Map {
			// Yes this is a map of [string]interface{}
			if vvalue.IsNil() {
				continue
			}
			nestedMap, ok := v.(map[string]interface{})
			if ok {
				traverseMap(nestedMap)
			}
		}

		// Is this a []interface{}?
		if vkind == reflect.Slice {
			// Yes.
			if vvalue.IsNil() {
				continue
			}
			nestedSlice, ok := v.([]interface{})
			if ok {
				traverseSlice(nestedSlice)
			}
		}
	}
}

// traverseSlice iterates through all values in a []interface{}. If it finds a
// nested map[string]interface{} or a []interface we recurse into it.
func traverseSlice(s []interface{}) {
	for _, v := range s {

		// Is this a map of [string]interface{}?
		vvalue := reflect.ValueOf(v)
		vkind := vvalue.Kind()
		if vkind == reflect.Map {
			// Yes this is a map [string]interface{}
			if vvalue.IsNil() {
				continue
			}
			nestedMap, ok := v.(map[string]interface{})
			if ok {
				traverseMap(nestedMap)
			}
		}

		// Is this a []interface{}
		if vkind == reflect.Slice {
			// Yes.
			if vvalue.IsNil() {
				continue
			}
			nestedSlice, ok := v.([]interface{})
			if ok {
				traverseSlice(nestedSlice)
			}
		}
	}
}
