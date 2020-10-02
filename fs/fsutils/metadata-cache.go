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

package fsutils

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
func (c *MetadataCache) Get(name string) (headerVersion uint16, headerLength int32, wrappedKey []byte, metadataLength int32, metadata *crypto.Metadata) {
	el, ok := c.cache.Get(name)
	if !ok {
		return
	}

	entry, ok := el.(*metadataCacheEntry)
	if !ok {
		return
	}

	headerVersion = entry.headerVersion
	headerLength = entry.headerLength
	wrappedKey = entry.wrappedKey
	metadataLength = entry.metadataLength
	metadata = entry.metadata

	return
}

// Add an item to the cache
func (c *MetadataCache) Add(name string, headerVersion uint16, headerLength int32, wrappedKey []byte, metadataLength int32, metadata *crypto.Metadata) {
	entry := &metadataCacheEntry{
		headerVersion:  headerVersion,
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

// Keys returns the list of keys in the cache
func (c *MetadataCache) Keys() []string {
	keys := c.cache.Keys()
	res := make([]string, len(keys))
	for i, el := range keys {
		res[i] = el.(string)
	}
	return res
}

// Remove an element from the cache
func (c *MetadataCache) Remove(name string) {
	c.cache.Remove(name)
}

// Purge the cache, removing all elements
func (c *MetadataCache) Purge() {
	c.cache.Purge()
}

type metadataCacheEntry struct {
	headerVersion                uint16
	headerLength, metadataLength int32
	wrappedKey                   []byte
	metadata                     *crypto.Metadata
}
