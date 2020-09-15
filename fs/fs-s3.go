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
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/utils"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Register the fs
func init() {
	t := reflect.TypeOf((*S3)(nil)).Elem()
	fsTypes["s3"] = t
	fsTypes["minio"] = t
}

// S3 stores files on a S3-compatible service
// This implementation does not rely on tags because S3 does not support conditional put requests
type S3 struct {
	fsBase

	client     *minio.Core
	bucketName string
	cache      *MetadataCache
	mux        sync.Mutex
}

func (f *S3) InitWithOptionsMap(opts map[string]string, cache *MetadataCache) error {
	// Required keys: "bucket", "accessKey", "secretKey"
	// Optional keys: "endpoint", "tls"

	// Load from the environment whatever we can (will be used as fallback
	f.loadEnvVars(opts)

	// Cache
	f.cache = cache

	// Bucket name
	if opts["bucket"] == "" {
		return errors.New("option 'bucket' is not defined")
	}
	f.bucketName = opts["bucket"]

	// Access key and secret key
	if opts["accessKey"] == "" || opts["secretKey"] == "" {
		return errors.New("options 'accessKey' and/or 'secretKey' are not defined")
	}

	// Endpoint
	// If not set, defaults to "s3.amazonaws.com"
	endpoint := opts["endpoint"]
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	}

	// Enable TLS
	// If not set, defaults to true
	tls := true
	tlsStr := strings.ToLower(opts["tls"])
	if tlsStr == "0" || tlsStr == "n" || tlsStr == "no" || tlsStr == "false" {
		tls = false
	}

	// Initialize minio client object for connecting to S3
	// Client is a higher-level API, that is convenient for things like putting files
	// Core is a lower-level API, which is easier for us when requesting data
	minioOpts := &minio.Options{
		Creds:  credentials.NewStaticV4(opts["accessKey"], opts["secretKey"], ""),
		Secure: tls,
	}
	var err error
	f.client, err = minio.NewCore(endpoint, minioOpts)
	if err != nil {
		return err
	}

	return nil
}

func (f *S3) InitWithConnectionString(connection string, cache *MetadataCache) error {
	opts := make(map[string]string)

	// Ensure the connection string is valid and extract the parts
	// connection must start with "s3:"
	// Then it must contain the bucket name, and optionally the access key and secret key (separated by : )
	// Lastly, might aso have the endpoint and whether to enable TLS
	parts := strings.Split(connection, ":")
	if len(parts) < 2 {
		return errors.New("invalid connection string")
	}
	opts["bucket"] = parts[1]

	// Check if we have the access key and secret key
	if len(parts) >= 4 {
		opts["accessKey"] = parts[2]
		opts["secretKey"] = parts[3]

		// Check if we have the endpoint
		if len(parts) >= 5 {
			opts["endpoint"] = parts[4]

			// Lastly, check if we have an option for TLS
			if len(parts) >= 6 {
				opts["tls"] = parts[5]
			}
		}
	}

	// Init the object from the opts dictionary
	return f.InitWithOptionsMap(opts, cache)
}

func (f *S3) loadEnvVars(opts map[string]string) {
	if opts["bucket"] == "" {
		opts["bucket"] = os.Getenv("S3_BUCKET")
	}
	if opts["accessKey"] == "" {
		opts["accessKey"] = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if opts["secretKey"] == "" {
		opts["secretKey"] = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if opts["endpoint"] == "" {
		opts["endpoint"] = os.Getenv("S3_ENDPOINT")
	}
	if opts["tls"] == "" {
		opts["tls"] = os.Getenv("S3_TLS")
	}
}

func (f *S3) AccountName() string {
	return f.bucketName
}

func (f *S3) GetInfoFile() (info *infofile.InfoFile, err error) {
	// Request the file from S3
	obj, _, _, err := f.client.GetObject(context.Background(), f.bucketName, "_info.json", minio.GetObjectOptions{})
	if err != nil {
		// Check if it's a minio error and it's a not found one
		e, ok := err.(minio.ErrorResponse)
		if ok && e.Code == "NoSuchKey" {
			err = nil
		}
		return
	}
	defer obj.Close()

	// Read the entire file
	data, err := ioutil.ReadAll(obj)
	if err != nil || len(data) == 0 {
		return
	}

	// Parse the JSON data
	info = &infofile.InfoFile{}
	if err = json.Unmarshal(data, info); err != nil {
		info = nil
		return
	}

	// Validate the content
	if err = info.Validate(); err != nil {
		info = nil
		return
	}

	// Set the data path
	f.dataPath = info.DataPath

	return
}

func (f *S3) SetInfoFile(info *infofile.InfoFile) (err error) {
	// Encode the content as JSON
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(data)

	// Upload the file
	_, err = f.client.Client.PutObject(context.Background(), f.bucketName, "_info.json", buf, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return err
	}

	return
}

func (f *S3) Get(ctx context.Context, name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if name[0] != '_' {
		folder = f.dataPath + "/"
	}

	found = true

	// Request the file from S3
	obj, stat, _, err := f.client.GetObject(ctx, f.bucketName, folder+name, minio.GetObjectOptions{})
	if err != nil {
		// Check if it's a minio error and it's a not found one
		e, ok := err.(minio.ErrorResponse)
		if ok && e.Code == "NoSuchKey" {
			err = nil
			found = false
		}
		return
	}
	defer obj.Close()

	// Check if the file exists but it's empty
	if stat.Size == 0 {
		found = false
		return
	}

	// Decrypt the data
	var metadataLength int32
	var metadata *crypto.Metadata
	headerVersion, headerLength, wrappedKey, err := crypto.DecryptFile(ctx, out, obj, f.masterKey, func(md *crypto.Metadata, sz int32) bool {
		metadata = md
		metadataLength = sz
		if metadataCb != nil {
			metadataCb(md, sz)
		}
		return true
	})
	// Ignore ErrMetadataOnly so the metadata is still added to cache
	if err != nil && err != crypto.ErrMetadataOnly {
		return
	}

	// Store the metadata in cache
	// Adding a lock here to prevent the case when adding this key causes the eviction of another one that's in use
	f.mux.Lock()
	f.cache.Add(name, headerVersion, headerLength, wrappedKey, metadataLength, metadata)
	f.mux.Unlock()

	return
}

func (f *S3) GetWithRange(ctx context.Context, name string, out io.Writer, rng *RequestRange, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if name[0] != '_' {
		folder = f.dataPath + "/"
	}

	found = true
	var obj io.ReadCloser
	var stat minio.ObjectInfo
	var opts minio.GetObjectOptions

	// Look up the file's metadata in the cache
	f.mux.Lock()
	headerVersion, headerLength, wrappedKey, metadataLength, metadata := f.cache.Get(name)
	if headerVersion == 0 || headerLength < 1 || wrappedKey == nil || len(wrappedKey) < 1 {
		// Need to request the metadata and cache it
		// For that, we need to request the header and the first package, which are at most 64kb + (32+256) bytes
		var length int64 = 64*1024 + 32 + 256
		innerCtx, cancel := context.WithCancel(ctx)

		// Request the file from S3
		opts = minio.GetObjectOptions{}
		opts.SetRange(0, length)
		obj, stat, _, err = f.client.GetObject(innerCtx, f.bucketName, folder+name, opts)
		if err != nil {
			// Check if it's a minio error and it's a not found one
			e, ok := err.(minio.ErrorResponse)
			if ok && e.Code == "NoSuchKey" {
				err = nil
				found = false
			}
			f.mux.Unlock()
			cancel()
			return
		}
		defer obj.Close()

		// Check if the file exists but it's empty
		if stat.Size == 0 {
			f.mux.Unlock()
			cancel()
			found = false
			return
		}

		// Decrypt the data
		headerVersion, headerLength, wrappedKey, err = crypto.DecryptFile(ctx, nil, obj, f.masterKey, func(md *crypto.Metadata, sz int32) bool {
			metadata = md
			metadataLength = sz
			cancel()
			return false
		})
		if err != nil && err != crypto.ErrMetadataOnly {
			f.mux.Unlock()
			cancel()
			return
		}

		// Store the metadata in cache
		f.cache.Add(name, headerVersion, headerLength, wrappedKey, metadataLength, metadata)
	}
	f.mux.Unlock()

	// Add the offsets to the range object and set the file size (it's guaranteed it's set, or we wouldn't have a range request)
	rng.HeaderOffset = int64(headerLength)
	rng.MetadataOffset = int64(metadataLength)
	rng.SetFileSize(metadata.Size)

	// Send the metadata to the callback
	if metadataCb != nil {
		metadataCb(metadata, metadataLength)
	}

	// Request the actual ranges that we need
	opts = minio.GetObjectOptions{}
	opts.SetRange(rng.StartBytes(), rng.EndBytes()-1)
	obj, stat, _, err = f.client.GetObject(ctx, f.bucketName, folder+name, opts)
	if err != nil {
		return
	}
	defer obj.Close()

	// Check if the file exists but it's empty
	if stat.Size == 0 {
		found = false
		return
	}

	// Decrypt the data
	err = crypto.DecryptPackages(ctx, out, obj, headerVersion, wrappedKey, f.masterKey, rng.StartPackage(), uint32(rng.SkipBeginning()), rng.Length, nil)
	if err != nil {
		return
	}

	return
}

func (f *S3) Set(ctx context.Context, name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return nil, err
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if name[0] != '_' {
		folder = f.dataPath + "/"
	}

	// Encrypt the data and upload it
	pr, pw := io.Pipe()
	var innerErr error
	go func() {
		defer pw.Close()
		r := utils.ReaderFuncWithContext(ctx, in)
		innerErr = crypto.EncryptFile(pw, r, f.masterKey, metadata)
	}()
	_, err = f.client.Client.PutObject(ctx, f.bucketName, folder+name, pr, -1, minio.PutObjectOptions{})
	if innerErr != nil {
		return nil, innerErr
	}
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (f *S3) Delete(ctx context.Context, name string, tag interface{}) (err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if name[0] != '_' {
		folder = f.dataPath + "/"
	}

	err = f.client.Client.RemoveObject(ctx, f.bucketName, folder+name, minio.RemoveObjectOptions{})

	return
}
