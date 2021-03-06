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
	"strconv"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs"
)

// IndexProviderFs provides access to the index using a fs.Fs back-end for storage
type IndexProviderFs struct {
	Store fs.Fs
}

// Get the index file
func (i *IndexProviderFs) Get(ctx context.Context, sequence uint32) (data []byte, isJSON bool, tag interface{}, err error) {
	// Abort if no store
	if i.Store == nil {
		err = errors.New("store is not initialized")
		return
	}

	// File name based on the sequence
	// The first doesn't have a suffix because of backwards-compatibility
	name := "_index"
	if sequence > 0 {
		name += "." + strconv.FormatUint(uint64(sequence), 10)
	}

	// Need to request the index
	buf := &bytes.Buffer{}
	var found bool
	found, tag, err = i.Store.Get(ctx, name, buf, func(metadata *crypto.Metadata, metadataSize int32) {
		// Check if we're decoding a legacy JSON file
		if metadata.ContentType == "application/json" {
			isJSON = true
		} else if metadata.ContentType != "application/protobuf" {
			err = errors.New("invalid Content-Type: " + metadata.ContentType)
		}
	})
	if found && buf.Len() > 0 {
		// Check error here because otherwise we might have an error also if the index wasn't found
		if err != nil {
			return
		}

		data = buf.Bytes()
	} else {
		// Ignore "not found" errors
		err = nil
	}

	return
}

// Set the index file
func (i *IndexProviderFs) Set(ctx context.Context, data []byte, sequence uint32, tag interface{}) (newTag interface{}, err error) {
	// Abort if no store
	if i.Store == nil {
		err = errors.New("store is not initialized")
		return
	}

	// Ensure data is not empty
	if len(data) == 0 {
		return nil, errors.New("data must not be empty")
	}

	// File name based on the sequence
	name := "_index"
	if sequence > 0 {
		name += "." + strconv.FormatUint(uint64(sequence), 10)
	}

	// Encrypt and save the updated index, if the tag is the same
	metadata := &crypto.Metadata{
		Name:        "index",
		ContentType: "application/protobuf",
		Size:        int64(len(data)),
	}
	buf := bytes.NewBuffer(data)
	return i.Store.Set(ctx, name, buf, tag, metadata)
}
