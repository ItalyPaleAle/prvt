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

package fsindex

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs"
)

// IndexProviderFs provides access to the index using a fs.Fs back-end for storage
type IndexProviderFs struct {
	Store fs.Fs
}

// Get the index file
func (i *IndexProviderFs) Get() (data []byte, isJSON bool, tag interface{}, err error) {
	isJSON = false

	// Abort if no store
	if i.Store == nil {
		err = errors.New("store is not initialized")
		return
	}

	// Need to request the index
	buf := &bytes.Buffer{}
	var found bool
	found, tag, err = i.Store.Get(context.Background(), "_index", buf, func(metadata *crypto.Metadata, metadataSize int32) {
		// Check if we're decoding a legacy JSON file
		if metadata.ContentType == "application/json" {
			isJSON = true
		} else if metadata.ContentType != "application/protobuf" {
			err = errors.New("invalid Content-Type: " + metadata.ContentType)
		}
	})
	if found {
		// Check error here because otherwise we might have an error also if the index wasn't found
		if err != nil {
			return
		}

		data, err = ioutil.ReadAll(buf)
		if err != nil {
			return
		}
	} else {
		// Ignore "not found" errors
		err = nil
	}

	return
}

// Set the index file
func (i *IndexProviderFs) Set(data []byte, cacheTag interface{}) (newTag interface{}, err error) {
	// Abort if no store
	if i.Store == nil {
		err = errors.New("store is not initialized")
		return
	}

	// Ensure data is not empty
	if data == nil || len(data) == 0 {
		return nil, errors.New("data must not be empty")
	}

	// Encrypt and save the updated index, if the tag is the same
	metadata := &crypto.Metadata{
		Name:        "index",
		ContentType: "application/protobuf",
		Size:        int64(len(data)),
	}
	buf := bytes.NewBuffer(data)
	return i.Store.Set(context.Background(), "_index", buf, cacheTag, metadata)
}
