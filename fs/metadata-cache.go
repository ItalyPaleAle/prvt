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
	"github.com/ItalyPaleAle/prvt/crypto"

	lru "github.com/hashicorp/golang-lru"
)

// Cache the metadata for at most 200 items (it's a LRU cache)
const metadataCacheSize = 200

// MetadataCache is a LRU cache for metadata objects
type MetadataCache struct {
	cache *lru.TwoQueueCache
}

// Init the object
func (c *MetadataCache) Init() (err error) {
	c.cache, err = lru.New2Q(metadataCacheSize)
	return
}

// Get returns an element from the cache
func (c *MetadataCache) Get(name string) (headerLength int32, wrappedKey []byte, metadataLength int32, metadata *crypto.Metadata) {
	el, ok := c.cache.Get(name)
	if !ok {
		return
	}

	entry, ok := el.(*metadataCacheEntry)
	if !ok {
		return
	}

	headerLength = entry.headerLength
	wrappedKey = entry.wrappedKey
	metadataLength = entry.metadataLength
	metadata = entry.metadata

	return
}

// Add an item to the cache
func (c *MetadataCache) Add(name string, headerLength int32, wrappedKey []byte, metadataLength int32, metadata *crypto.Metadata) {
	entry := &metadataCacheEntry{
		headerLength:   headerLength,
		metadataLength: metadataLength,
		wrappedKey:     wrappedKey,
		metadata:       metadata,
	}
	c.cache.Add(name, entry)
}

// Contains returns true if the item is cached
func (c *MetadataCache) Contains(name string) bool {
	return c.cache.Contains(name)
}

// Remove an element from the cache
func (c *MetadataCache) Remove(name string) {
	c.cache.Remove(name)
}

type metadataCacheEntry struct {
	headerLength, metadataLength int32
	wrappedKey                   []byte
	metadata                     *crypto.Metadata
}
