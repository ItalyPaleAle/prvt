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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 stores files on a S3-compatible service
// This implementation does not rely on tags because S3 does not support conditional put requests
type S3 struct {
	masterKey  []byte
	client     *minio.Client
	core       *minio.Core
	bucketName string
	dataPath   string
	cache      *MetadataCache
	mux        sync.Mutex
}

func (f *S3) Init(connection string, cache *MetadataCache) error {
	f.cache = cache

	// Ensure the connection string is valid and extract the parts
	// connection must start with "s3:"
	// Then it must contain the bucket name
	if !strings.HasPrefix(connection, "s3:") || len(connection) < 4 {
		return fmt.Errorf("invalid scheme")
	}
	f.bucketName = connection[3:]

	// Get the access key
	accessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKeyId == "" || secretAccessKey == "" {
		return errors.New("environmental variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY are not defined")
	}

	// Endpoint
	// If not set, defaults to "s3.amazonaws.com"
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	}

	// Enable TLS
	// If not set, defaults to true
	tls := true
	tlsStr := strings.ToLower(os.Getenv("S3_TLS"))
	if tlsStr == "0" || tlsStr == "n" || tlsStr == "no" || tlsStr == "false" {
		tls = false
	}

	// Initialize minio client object for connecting to S3
	// Client is a higher-level API, that is convenient for things like putting files
	// Core is a lower-level API, which is easier for us when requesting data
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretAccessKey, ""),
		Secure: tls,
	}
	var err error
	f.client, err = minio.New(endpoint, opts)
	if err != nil {
		return err
	}
	f.core, err = minio.NewCore(endpoint, opts)
	if err != nil {
		return err
	}

	return nil
}

func (f *S3) SetDataPath(path string) {
	f.dataPath = path
}

func (f *S3) SetMasterKey(key []byte) {
	f.masterKey = key
}

func (f *S3) GetInfoFile() (info *infofile.InfoFile, err error) {
	// Request the file from S3
	obj, _, _, err := f.core.GetObject(context.Background(), f.bucketName, "_info.json", minio.GetObjectOptions{})
	if err != nil {
		return nil, err
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
	_, err = f.client.PutObject(context.Background(), f.bucketName, "_info.json", buf, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return err
	}

	return
}

func (f *S3) Get(name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
	return f.GetWithContext(context.Background(), name, out, metadataCb)
}

func (f *S3) GetWithContext(ctx context.Context, name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
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
	obj, stat, _, err := f.core.GetObject(ctx, f.bucketName, folder+name, minio.GetObjectOptions{})
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
	var metadataLength int32
	var metadata *crypto.Metadata
	headerLength, wrappedKey, err := crypto.DecryptFile(out, obj, f.masterKey, func(md *crypto.Metadata, sz int32) {
		metadata = md
		metadataLength = sz
		metadataCb(md, sz)
	})
	if err != nil {
		return
	}

	// Store the metadata in cache
	// Adding a lock here to prevent the case when adding this key causes the eviction of another one that's in use
	f.mux.Lock()
	f.cache.Add(name, headerLength, wrappedKey, metadataLength, metadata)
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
	headerLength, wrappedKey, metadataLength, metadata := f.cache.Get(name)
	if headerLength < 1 || wrappedKey == nil || len(wrappedKey) < 1 {
		// Need to request the metadata and cache it
		// For that, we need to request the header and the first package, which are at most 64kb + (32+256) bytes
		var length int64 = 64*1024 + 32 + 256
		innerCtx, cancel := context.WithCancel(ctx)

		// Request the file from S3
		opts = minio.GetObjectOptions{}
		opts.SetRange(0, length)
		obj, stat, _, err = f.core.GetObject(innerCtx, f.bucketName, folder+name, opts)
		if err != nil {
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
		headerLength, wrappedKey, err = crypto.DecryptFile(nil, obj, f.masterKey, func(md *crypto.Metadata, sz int32) {
			metadata = md
			metadataLength = sz
			cancel()
		})
		if err != nil && err != crypto.ErrMetadataOnly {
			f.mux.Unlock()
			cancel()
			return
		}

		// Store the metadata in cache
		f.cache.Add(name, headerLength, wrappedKey, metadataLength, metadata)
	}
	f.mux.Unlock()

	// Add the offsets to the range object and set the file size (it's guaranteed it's set, or we wouldn't have a range request)
	rng.HeaderOffset = int64(headerLength)
	rng.MetadataOffset = int64(metadataLength)
	rng.SetFileSize(metadata.Size)

	// Send the metadata to the callback
	metadataCb(metadata, metadataLength)

	// Request the actual ranges that we need
	opts = minio.GetObjectOptions{}
	opts.SetRange(rng.StartBytes(), rng.EndBytes()-1)
	obj, stat, _, err = f.core.GetObject(ctx, f.bucketName, folder+name, opts)
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
	err = crypto.DecryptPackages(out, obj, wrappedKey, f.masterKey, rng.StartPackage(), uint32(rng.SkipBeginning()), rng.Length, nil)
	if err != nil {
		return
	}

	return
}

func (f *S3) Set(name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error) {
	return f.SetWithContext(context.Background(), name, in, tag, metadata)
}

func (f *S3) SetWithContext(ctx context.Context, name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error) {
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
	go func() {
		err := crypto.EncryptFile(pw, in, f.masterKey, metadata)
		if err != nil {
			panic(err)
		}
		pw.Close()
	}()
	_, err = f.client.PutObject(ctx, f.bucketName, folder+name, pr, -1, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (f *S3) Delete(name string, tag interface{}) (err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if name[0] != '_' {
		folder = f.dataPath + "/"
	}

	err = f.client.RemoveObject(context.Background(), f.bucketName, folder+name, minio.RemoveObjectOptions{})

	return
}
