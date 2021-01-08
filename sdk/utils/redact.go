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

package utils

import (
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
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
func RedactPasswords(obj interface{}) (redacted interface{}, err error) {
	// Wrap the original obj in a reflect.Value
	original := reflect.ValueOf(obj)

	if !original.IsValid() {
		return obj, nil
	}

	// Create a new copy of the object type
	copied := reflect.New(original.Type()).Elem()
	err = redactRecursive(copied, original)
	if err != nil {
		return
	}

	// Return the value of the copy
	return copied.Interface(), nil
}

func redactRecursive(copied, original reflect.Value) (err error) {
	switch original.Kind() {

	// If a pointer, unwrap and call again.
	case reflect.Ptr:
		originalValue := original.Elem()
		// Check if the pointer is nil. If so, there is nothing to do here.
		if !originalValue.IsValid() {
			return
		}
		// Create a new object and set the pointer to it, then recurse.
		copied.Set(reflect.New(originalValue.Type()))
		err = redactRecursive(copied.Elem(), originalValue)
		if err != nil {
			return
		}

	// If an interface, unwrap the interface and recurse.
	case reflect.Interface:
		// Unwrap the interface
		originalValue := original.Elem()
		// Create a new object. New gives us a pointer which we don't want,
		// so call Elem to dereference the pointer.
		if !originalValue.IsValid() {
			// This is an untyped nil interface. We cannot set to the originalValue.
			// See https://github.com/vapor-ware/synse-sdk/issues/480.
			//

			//copied.Set(originalValue)
		} else {
			copyValue := reflect.New(originalValue.Type()).Elem()
			err = redactRecursive(copyValue, originalValue)
			if err != nil {
				return
			}
			copied.Set(copyValue)
		}

	// If a slice, create a new slice and check each element in the slice.
	case reflect.Slice:
		copied.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			err = redactRecursive(copied.Index(i), original.Index(i))
			if err != nil {
				return
			}
		}

	// If a map, create a new map and check each element in the map.
	case reflect.Map:
		copied.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			log.Debugf("Traversing map key: %v", key)
			originalValue := original.MapIndex(key)

			// First, check that the key is a string, and if so, that it contains the
			// "pass" substring. If the key is an interface, first unwrap the interface.
			if key.Kind() == reflect.Interface {
				key = key.Elem()
			}

			if key.Kind() == reflect.String {
				if strings.Contains(strings.ToLower(key.Interface().(string)), "pass") {
					// Check that the original value is a string or interface. If either
					// case is true, set the value to "REDACTED"
					switch originalValue.Kind() {
					case reflect.String:
						if !originalValue.IsValid() {
							copied.SetMapIndex(key, originalValue)
						} else {
							copyValue := reflect.New(originalValue.Type()).Elem()
							copyValue.SetString("REDACTED")
							copied.SetMapIndex(key, copyValue)
						}

					case reflect.Interface:
						if !originalValue.IsValid() {
							copied.SetMapIndex(key, originalValue)
						} else {
							copyValue := reflect.New(originalValue.Type()).Elem()
							copyValue.Set(reflect.ValueOf("REDACTED"))
							copied.SetMapIndex(key, copyValue)
						}
					}
				} else {
					if !originalValue.IsValid() {
						copied.SetMapIndex(key, originalValue)
					} else {
						copyValue := reflect.New(originalValue.Type()).Elem()
						err = redactRecursive(copyValue, originalValue)
						if err != nil {
							return
						}
						copied.SetMapIndex(key, copyValue)
					}
				}
			} else {
				copyValue := reflect.New(originalValue.Type()).Elem()
				err = redactRecursive(copyValue, originalValue)
				if err != nil {
					return
				}
				copied.SetMapIndex(key, copyValue)
			}
		}

	// Otherwise, simply take the original value.
	default:
		copied.Set(original)
	}
	return
}
