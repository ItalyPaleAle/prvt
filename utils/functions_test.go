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
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/unicode/norm"
)

func TestSanitizePath(t *testing.T) {
	var out string
	// Test from a table
	table := []struct {
		In     string
		Expect string
	}{
		// Simple
		{"foo", "foo"},
		// Normalization in NFKC form
		{"è", norm.NFKC.String("è")},
		{"ﬁ", norm.NFKC.String("ﬁ")},
		{"ñ", norm.NFKC.String("ñ")},
		// Remove invalid (not safe for URLs or filesystems) characters
		{"aa&", "aa"},
		{"aa=", "aa"},
		{"aa{}<>*", "aa"},
		// Replace back slashes with a forward one
		{"foo\\bar", "foo/bar"},
		{"foo\\bar\\2", "foo/bar/2"},
		// Replace multiple slashes with a single one
		{"//aa", "/aa"},
		{"hello////world", "hello/world"},
		{"hello////world//aa", "hello/world/aa"},
	}

	for _, el := range table {
		out = SanitizePath(el.In)
		assert.Equal(t, el.Expect, out)
	}
}

func TestIsTruthy(t *testing.T) {
	var out bool
	// Test from a table
	table := []struct {
		In     string
		Expect bool
	}{
		// True
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"TrUe", true},
		{"TrUe", true},
		{"t", true},
		{"T", true},
		{"y", true},
		{"Y", true},
		{"yes", true},
		{"YES", true},
		{"YeS", true},
		{" 1", true},
		{"1 ", true},
		{" 1 ", true},
		{" TruE ", true},
		{" T ", true},
		// Everything else is false
		{"0", false},
		{"false", false},
		{"f", false},
		{"no", false},
		{"n", false},
		{"N", false},
		{"hello world", false},
		{"not true", false},
		{"t rue", false},
		{"!", false},
		{"", false},
	}

	for _, el := range table {
		out = IsTruthy(el.In)
		assert.Equal(t, el.Expect, out)
	}
}
