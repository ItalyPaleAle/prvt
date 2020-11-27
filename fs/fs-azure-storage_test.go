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
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/ItalyPaleAle/prvt/fs/fsutils"
	"github.com/stretchr/testify/assert"
)

func TestFsAzure(t *testing.T) {
	// Ensure we have the credentials
	if os.Getenv("AZURE_STORAGE_ACCOUNT") == "" || os.Getenv("AZURE_STORAGE_ACCESS_KEY") == "" {
		t.Skip("missing Azure Storage credentials in environment")
		return
	}

	// Init the caches
	cache := &fsutils.MetadataCache{}
	err := cache.Init()
	if !assert.NoError(t, err) {
		return
	}

	// Generate a container name
	container := "prvttest" + RandString(6)

	// Init the object
	store := &AzureStorage{}
	opts := map[string]string{
		"type":      "azure",
		"container": container,
	}
	err = store.InitWithOptionsMap(opts, cache)
	if !assert.NoError(t, err) {
		return
	}

	// Create the container
	u, err := url.Parse(store.storageURL)
	if !assert.NoError(t, err) {
		return
	}
	containerUrl := azblob.NewContainerURL(*u, store.storagePipeline)
	_, err = containerUrl.Create(context.Background(), azblob.Metadata{}, azblob.PublicAccessNone)
	if !assert.NoError(t, err) {
		return
	}
	t.Log("Created container", container)
	defer removeAzureContainer(t, store, containerUrl)

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

func removeAzureContainer(t *testing.T, store *AzureStorage, containerUrl azblob.ContainerURL) {
	_, err := containerUrl.Delete(context.Background(), azblob.ContainerAccessConditions{})
	if !assert.NoError(t, err) {
		return
	}
	t.Log("Deleted container", containerUrl.String())
}
