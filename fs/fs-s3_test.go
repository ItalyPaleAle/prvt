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
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

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

	// Create 2 store objects
	store := &S3{}
	store2 := &S3{}
	opts := map[string]string{
		"type":   "s3",
		"bucket": bucket,
	}
	err = store.InitWithOptionsMap(opts, cache)
	if !assert.NoError(t, err) {
		return
	}
	err = store2.InitWithOptionsMap(opts, cache)
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

	// Test S3 locks
	t.Run("test S3 locks", testS3Locks(store, bucket))

	// Run the tests
	t.Run("common tests", func(t *testing.T) {
		tester := &testFs{
			t:      t,
			store:  store,
			store2: store2,
			cache:  cache,
		}
		tester.Run()
	})

	// At the end, there should be no locks left over
	t.Run("check lock folder", s3LockFolderPostCheck(store, bucket))
}

// For S3, locks are managed in a different way, so we need to do some special tests
func testS3Locks(store *S3, bucket string) func(t *testing.T) {
	return func(t *testing.T) {
		var (
			ctx = context.Background()
			err error
		)

		// Create 2 locks, waiting 4 seconds in-between
		_, err = store.client.Client.PutObject(ctx, bucket, "_locks/1", bytes.NewBuffer([]byte{1}), 1, minio.PutObjectOptions{})
		if !assert.NoError(t, err) {
			return
		}
		// Wait 5 seconds
		// (Sadly, we must wait and pause the test because Minio doesn't let us fake the LastModified times)
		time.Sleep(5 * time.Second)
		_, err = store.client.Client.PutObject(ctx, bucket, "_locks/2", bytes.NewBuffer([]byte{1}), 1, minio.PutObjectOptions{})
		if !assert.NoError(t, err) {
			return
		}

		// Set the lock duration to 3 seconds then try unlocking - call should return with error
		// Note: we need to bring down S3LockRenewal too
		prevS3LockDuration := S3LockDuration
		prevS3LockRenewal := S3LockRenewal
		S3LockDuration = 3
		S3LockRenewal = 2
		err = store.AcquireLock(ctx)
		if !assert.Error(t, err) {
			t.FailNow()
		}
		assert.Empty(t, store.lockId)
		assert.Empty(t, store.lockRefreshStop)

		// Sleep for 4 seconds so all locks expire, then try unlocking again - should succeed (and delete all expired locks)
		time.Sleep(4 * time.Second)
		err = store.AcquireLock(ctx)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		// Sleep for another 2 seconds to ensure older locks are removed and the refresher has started
		time.Sleep(2 * time.Second)
		assert.NotEmpty(t, store.lockId)
		assert.NotEmpty(t, store.lockRefreshStop)

		// Read the contents of the lock file
		lockContents1 := getS3LockFileContent(t, store, bucket)
		// Sleep for 2 seconds so the lock should be renewed (2 seconds before the expiration in 3 seconds)
		time.Sleep(2 * time.Second)
		// Read the contents of the lock again
		lockContents2 := getS3LockFileContent(t, store, bucket)
		// Contents should be different
		assert.True(t, lockContents1 != lockContents2)

		// Release the lock
		err = store.ReleaseLock(ctx)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.Empty(t, store.lockId)
		assert.Empty(t, store.lockRefreshStop)

		// Restore lock duration to the previous value
		S3LockDuration = prevS3LockDuration
		S3LockRenewal = prevS3LockRenewal
	}
}

func getS3LockFileContent(t *testing.T, store *S3, bucket string) string {
	t.Helper()

	ctx := context.Background()

	// List the locks (there should be only one)
	listCh := store.client.Client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix: "/_locks/",
	})
	lockName := ""
	for e := range listCh {
		// If there's already a name, means there was more than 1 file
		// We need to have a single lock
		assert.Empty(t, lockName, "found more than one lock file")
		lockName = e.Key
	}

	// Read the contents of the lock file
	r, _, _, err := store.client.GetObject(ctx, bucket, lockName, minio.GetObjectOptions{})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	content, err := ioutil.ReadAll(r)
	r.Close()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NotEmpty(t, content) {
		t.FailNow()
	}

	// Return the contents
	return string(content)
}

func s3LockFolderPostCheck(store *S3, bucket string) func(t *testing.T) {
	return func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)

		listCh := store.client.Client.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{
			Prefix: "/_locks/",
		})
		found := 0
		for range listCh {
			found++
		}
		assert.Zero(t, found, "found locks not correctly deleted")
	}
}

func removeS3Bucket(t *testing.T, store *S3, bucket string) {
	t.Helper()

	ctx := context.Background()

	// Delete all files first
	objectsCh := store.client.Client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Recursive: true,
	})
	deleteCh := store.client.Client.RemoveObjects(ctx, bucket, objectsCh, minio.RemoveObjectsOptions{})
	for e := range deleteCh {
		t.Log("Deleted object", e.ObjectName)
	}

	// Delete the bucket
	err := store.client.RemoveBucket(ctx, bucket)
	if err != nil {
		t.Errorf("error while removing the bucket %s: %s", bucket, err)
		return
	}

	t.Log("Deleted bucket", bucket)
}
