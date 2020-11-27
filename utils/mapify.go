/*
Copyright © 2020 Alessandro Segala (@ItalyPaleAle)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package utils

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
)

// Mapify converts a struct to a map[string]interface{}
// If fields are structs, they're converted to their textual representation
func Mapify(m interface{}) map[string]interface{} {
	// Get a reflection for this object
	// If it's a pointer, get the element it's pointing to
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	// Result
	result := make(map[string]interface{}, v.NumField())

	// Used for checking if a struct implements encoding.TextMarshaler
	textMarshalerType := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

	// Iterate through the fields of the struct
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// Get the field
		fieldTyp := typ.Field(i)
		fieldVal := v.Field(i)

		// Check if there's a `json` tag
		name := fieldTyp.Tag.Get("json")
		if name != "" {
			// Remove everything after , if present
			pos := strings.Index(name, ",")
			if pos >= 0 {
				name = name[0:pos]
			}
		}
		// If name is empty, it will be the same as the field's name
		// There's another if here in case the previous block made name empty
		if name == "" {
			name = fieldTyp.Name
		}

		// If the value is a pointer, get the element
		if fieldVal.Kind() == reflect.Ptr {
			fieldVal = reflect.Indirect(fieldVal)
			// Check if it's a pointer to nil or anything else invalid
			if !fieldVal.IsValid() {
				result[name] = nil
				continue
			}
		}

		// If this is a struct, convert to string
		if fieldVal.Kind() == reflect.Struct {
			// First, check if the struct implements encoding.TextMarshaler for marshaling
			// Otherwise, fall back to using json.Marshal
			if fieldTyp.Type.Implements(textMarshalerType) {
				txt, err := fieldVal.Interface().(encoding.TextMarshaler).MarshalText()
				if txt == nil || err != nil {
					// Ignore errors, and convert to an empty byte slice
					txt = []byte{}
				}
				result[name] = string(txt)
			} else {
				b, err := json.Marshal(fieldVal.Interface())
				if b == nil || err != nil {
					// Ignore errors, and convert to an empty byte slice
					b = []byte{}
				}
				result[name] = string(b)
			}
		} else if b, ok := fieldVal.Interface().([]byte); ok {
			// This is a byte slice, so encode it as base64 for consistency with json.Marshal
			result[name] = base64.StdEncoding.EncodeToString(b)
		} else {
			// Set the field in the result map
			result[name] = fieldVal.Interface()
		}
	}
	return result
}
