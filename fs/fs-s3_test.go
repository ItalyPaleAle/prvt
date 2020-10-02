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
	"os"
	"testing"

	"github.com/ItalyPaleAle/prvt/fs/fsutils"
	minio "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestFsS3(t *testing.T) {
	// Ensure we have the credentials
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		t.Skip("missing AWS credentials in environment")
		return
	}

	// Init the caches
	cache := &fsutils.MetadataCache{}
	err := cache.Init()
	if !assert.NoError(t, err) {
		return
	}

	// Generate a bucket name and get the region
	bucket := "prvttest" + RandString(6)
	region := os.Getenv("S3_REGION")

	// Init the object
	store := &S3{}
	opts := map[string]string{
		"type":   "s3",
		"bucket": bucket,
	}
	err = store.InitWithOptionsMap(opts, cache)
	if !assert.NoError(t, err) {
		return
	}

	// Create the bucket
	err = store.client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{
		Region: region,
	})
	if !assert.NoError(t, err) {
		return
	}
	t.Log("Created bucket", bucket)
	defer removeS3Bucket(t, store, bucket)

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

func removeS3Bucket(t *testing.T, store *S3, bucket string) {
	// Delete all files first
	objectsCh := store.client.Client.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{
		Recursive: true,
	})
	deleteCh := store.client.Client.RemoveObjects(context.Background(), bucket, objectsCh, minio.RemoveObjectsOptions{})
	for e := range deleteCh {
		t.Log("Deleted object", e.ObjectName)
	}

	// Delete the bucket
	err := store.client.RemoveBucket(context.Background(), bucket)
	if err != nil {
		t.Errorf("error while removing the bucket %s: %s", bucket, err)
		return
	}

	t.Log("Deleted bucket", bucket)
}
