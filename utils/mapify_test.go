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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMapify(t *testing.T) {
	var (
		res    map[string]interface{}
		expect map[string]interface{}
	)
	now := time.Now()
	nowBytes, _ := now.MarshalText()
	nowString := string(nowBytes)
	str := "bar"

	// Test a simple struct
	el1 := struct {
		Foo  string `json:"foo"`
		Num  int    `json:"answer,omitempty"` // ",omitempty" should be removed
		Bool bool
	}{
		Foo:  "bar",
		Num:  42,
		Bool: false,
	}
	expect = map[string]interface{}{
		"foo":    "bar",
		"answer": 42,
		"Bool":   false,
	}
	res = Mapify(el1)
	assert.True(t, reflect.DeepEqual(res, expect))

	// Pointer to a struct
	res = Mapify(&el1)
	assert.True(t, reflect.DeepEqual(res, expect))

	// Struct with time objects and nulls
	el2 := struct {
		Foo  string      `json:"foo"`
		Time time.Time   `json:"time"`
		Null interface{} `json:"null"`
	}{
		Foo:  "bar",
		Time: now,
		Null: nil,
	}
	expect = map[string]interface{}{
		"foo":  "bar",
		"time": string(nowString),
		"null": nil,
	}
	res = Mapify(el2)
	assert.True(t, reflect.DeepEqual(res, expect))

	// Struct with pointers inside
	el3 := struct {
		Foo      *string    `json:"foo"`
		Time     *time.Time `json:"time"`
		NullTime *time.Time `json:"nullTime"`
	}{
		Foo:      &str,
		Time:     &now,
		NullTime: nil,
	}
	expect = map[string]interface{}{
		"foo":      "bar",
		"time":     string(nowString),
		"nullTime": nil,
	}
	res = Mapify(el3)
	assert.True(t, reflect.DeepEqual(res, expect))

	// Struct with byte slices inside
	el4 := struct {
		Data []byte `json:"data"`
	}{
		Data: []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x21},
	}
	expect = map[string]interface{}{
		"data": "SGVsbG8h",
	}
	res = Mapify(el4)
	assert.True(t, reflect.DeepEqual(res, expect))
}
