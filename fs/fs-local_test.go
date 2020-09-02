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

package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFsLocal(t *testing.T) {
	// Temp directory
	tempDir := t.TempDir()

	// Init the cache
	cache := &MetadataCache{}
	err := cache.Init()
	if !assert.NoError(t, err) {
		return
	}

	// Init the object
	store := &Local{}
	opts := map[string]string{
		"type": "local",
		"path": tempDir,
	}
	err = store.InitWithOptionsMap(opts, cache)
	if !assert.NoError(t, err) {
		return
	}

	// Run the tests
	t.Run("common tests", func(t *testing.T) {
		tester := &testFs{
			t:     t,
			store: store,
			cache: cache,
		}
		tester.Run()
	})
}
