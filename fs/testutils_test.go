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
	"encoding/base64"
	"io"
	"io/ioutil"
	"math/rand"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/keys"

	"github.com/stretchr/testify/assert"
)

func init() {
	// Init the rng
	rand.Seed(time.Now().UnixNano())
}

// Performs tests for a store object, already initialized
type testFs struct {
	t     *testing.T
	store Fs
	cache *MetadataCache

	info  *infofile.InfoFile
	files map[string][]byte
}

// Starts the test
func (s *testFs) Run() {
	// Load fixtures
	s.loadFixtures()

	// Initialize repo
	s.testGetInfoFileNotInitialized()
	s.testSetInfoFile()
	s.testGetInfoFile()

	// Set and get raw files
	s.testRawSetFile()
	s.testRawGetFile()

	// Derive and set master key
	masterKey, keyId, _, err := keys.GetMasterKeyWithPassphrase(s.info, "hello world")
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	s.store.SetMasterKey(keyId, masterKey)

	// Set, retrieve, and delete encrypted files
	s.testSet()
	s.testGet()
	s.testGetWithRange()
	s.testDelete()
}

// Loads fixtures
func (s *testFs) loadFixtures() {
	// Info file
	s.info = staticInfoFile()

	// Sample files
	s.files = map[string][]byte{}
	read, err := ioutil.ReadFile(path.Join("..", "tests", "fixtures", "divinacommedia.txt"))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	s.files["divinacommedia.txt"] = read
	read, err = ioutil.ReadFile(path.Join("..", "tests", "fixtures", "short.txt"))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	s.files["short.txt"] = read
	read, err = ioutil.ReadFile(path.Join("..", "tests", "fixtures", "kitera-dent-BIj4LObC6es-unsplash.jpg"))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	s.files["kitera-dent-BIj4LObC6es-unsplash.jpg"] = read
}

// Get info file, but repo is not initialized
func (s *testFs) testGetInfoFileNotInitialized() {
	// Must have no error but nil info, meaning the file was not found
	info, err := s.store.GetInfoFile()
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.Nil(s.t, info)
}

// Set info file
func (s *testFs) testSetInfoFile() {
	err := s.store.SetInfoFile(s.info)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
}

// Get info file
func (s *testFs) testGetInfoFile() {
	info, err := s.store.GetInfoFile()
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.NotNil(s.t, info)
	assert.True(s.t, reflect.DeepEqual(info, s.info))
}

// Set raw files
func (s *testFs) testRawSetFile() {
	in := bytes.NewBufferString("hello world!")
	_, err := s.store.RawSet(context.Background(), "_raw", in, nil)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
}

// Get raw files
func (s *testFs) testRawGetFile() {
	var (
		found bool
		out   *bytes.Buffer
		err   error
	)

	// Entire file
	out = &bytes.Buffer{}
	found, _, err = s.store.RawGet(context.Background(), "_raw", out, 0, 0)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.True(s.t, found)
	assert.Equal(s.t, "hello world!", out.String())

	// Partial request - till the end
	out = &bytes.Buffer{}
	found, _, err = s.store.RawGet(context.Background(), "_raw", out, 6, 0)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.True(s.t, found)
	assert.Equal(s.t, "world!", out.String())

	// Partial request
	out = &bytes.Buffer{}
	found, _, err = s.store.RawGet(context.Background(), "_raw", out, 6, 2)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.True(s.t, found)
	assert.Equal(s.t, "wo", out.String())
}

// Store encrypted files
func (s *testFs) testSet() {
	var (
		in       io.Reader
		metadata *crypto.Metadata
		err      error
	)

	// Store text file
	in = bytes.NewReader(s.files["divinacommedia.txt"])
	metadata = &crypto.Metadata{
		Name:        "divinacommedia.txt",
		ContentType: "text/plain",
		Size:        int64(len(s.files["divinacommedia.txt"])),
	}
	_, err = s.store.Set(context.Background(), "divinacommedia.txt", in, nil, metadata)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}

	// Error: empty name
	_, err = s.store.Set(context.Background(), "", in, nil, metadata)
	if !assert.Error(s.t, err) {
		s.t.FailNow()
	}

	// Store image
	in = bytes.NewReader(s.files["kitera-dent-BIj4LObC6es-unsplash.jpg"])
	metadata = &crypto.Metadata{
		Name:        "kitera-dent-BIj4LObC6es-unsplash.jpg",
		ContentType: "image/jpeg",
		Size:        int64(len(s.files["kitera-dent-BIj4LObC6es-unsplash.jpg"])),
	}
	_, err = s.store.Set(context.Background(), "kitera-dent-BIj4LObC6es-unsplash.jpg", in, nil, metadata)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}

	// Context canceled
	{
		ctx, cancel := context.WithCancel(context.Background())
		f := bytes.NewReader(s.files["kitera-dent-BIj4LObC6es-unsplash.jpg"])
		pr, pw := io.Pipe()
		go func() {
			// Copy the first 10kb, then cancel the context, then copy the rest
			io.CopyN(pw, f, 10240)
			cancel()
			io.Copy(pw, f)
		}()
		//_, err = s.store.Set(ctx, "void", pr, nil, metadata)
		// This overwrites an existing file, but it should fail without actually overwriting it
		_, err = s.store.Set(ctx, "kitera-dent-BIj4LObC6es-unsplash.jpg", pr, nil, metadata)
		if !assert.Error(s.t, err) ||
			!assert.Equal(s.t, context.Canceled, err) {
			s.t.FailNow()
		}
	}

	// Store short text file
	in = bytes.NewReader(s.files["short.txt"])
	metadata = &crypto.Metadata{
		Name:        "short.txt",
		ContentType: "text/plain",
		Size:        int64(len(s.files["short.txt"])),
	}
	_, err = s.store.Set(context.Background(), "short.txt", in, nil, metadata)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
}

// Retrieve encrypted files
func (s *testFs) testGet() {
	var (
		out      io.Writer
		found    bool
		err      error
		read     []byte
		cbCalled bool
	)

	// Cache must be empty at this point
	assert.Len(s.t, s.cache.Keys(), 0)

	// Retrieve the text file
	out = &bytes.Buffer{}
	cbCalled = false
	found, _, err = s.store.Get(context.Background(), "divinacommedia.txt", out, func(metadata *crypto.Metadata, metadataSize int32) {
		assert.Equal(s.t, "divinacommedia.txt", metadata.Name)
		assert.Equal(s.t, "text/plain", metadata.ContentType)
		assert.Equal(s.t, int64(len(s.files["divinacommedia.txt"])), metadata.Size)
		cbCalled = true
	})
	if !assert.NoError(s.t, err) ||
		!assert.True(s.t, found) ||
		!assert.True(s.t, cbCalled) {
		s.t.FailNow()
	}
	read, err = ioutil.ReadAll(out.(io.ReadWriter))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.True(s.t, bytes.Equal(read, s.files["divinacommedia.txt"]))

	// File's metadata must be cached now
	s.checkCacheTextFile()

	// Retrieve the metadata only
	cbCalled = false
	found, _, err = s.store.Get(context.Background(), "kitera-dent-BIj4LObC6es-unsplash.jpg", nil, func(metadata *crypto.Metadata, metadataSize int32) {
		assert.Equal(s.t, "kitera-dent-BIj4LObC6es-unsplash.jpg", metadata.Name)
		assert.Equal(s.t, "image/jpeg", metadata.ContentType)
		assert.Equal(s.t, int64(len(s.files["kitera-dent-BIj4LObC6es-unsplash.jpg"])), metadata.Size)
		cbCalled = true
	})
	if !assert.EqualError(s.t, err, crypto.ErrMetadataOnly.Error()) ||
		!assert.True(s.t, found) ||
		!assert.True(s.t, cbCalled) {
		s.t.FailNow()
	}

	// Second file's metadata must be cached now
	assert.Len(s.t, s.cache.Keys(), 2)
	assert.True(s.t, s.cache.Contains("kitera-dent-BIj4LObC6es-unsplash.jpg"))

	// Error: empty name
	_, _, err = s.store.Get(context.Background(), "", nil, nil)
	if !assert.Error(s.t, err) {
		s.t.FailNow()
	}

	// File does not exist
	found, _, err = s.store.Get(context.Background(), "no_exist", nil, nil)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.False(s.t, found)

	// No metadata callback
	found, _, err = s.store.Get(context.Background(), "short.txt", nil, nil)
	if !assert.EqualError(s.t, err, crypto.ErrMetadataOnly.Error()) ||
		!assert.True(s.t, found) {
		s.t.FailNow()
	}

	// Third file's metadata must be cached now
	assert.Len(s.t, s.cache.Keys(), 3)
	assert.True(s.t, s.cache.Contains("short.txt"))

	// Short file
	out = &bytes.Buffer{}
	found, _, err = s.store.Get(context.Background(), "short.txt", out, nil)
	if !assert.NoError(s.t, err) ||
		!assert.True(s.t, found) {
		s.t.FailNow()
	}
	read, err = ioutil.ReadAll(out.(io.ReadWriter))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.True(s.t, bytes.Equal(read, s.files["short.txt"]))

	// Context canceled
	{
		cbCalled = false
		ctx, cancel := context.WithCancel(context.Background())
		pr, pw := io.Pipe()
		go func() {
			// Read the first 10 bytes, then cancel the context, then copy the rest
			buf := make([]byte, 10)
			io.ReadFull(pr, buf)
			cancel()
			io.Copy(ioutil.Discard, pr)
		}()
		_, _, err = s.store.Get(ctx, "kitera-dent-BIj4LObC6es-unsplash.jpg", pw, func(metadata *crypto.Metadata, metadataSize int32) {
			cbCalled = (metadataSize > 0)
		})
		if !assert.Error(s.t, err) ||
			!assert.Equal(s.t, context.Canceled, err) ||
			!assert.True(s.t, cbCalled) {
			s.t.FailNow()
		}
	}
}

// Retrieve partial encrypted files
func (s *testFs) testGetWithRange() {
	// Empty the metadata cache
	s.cache.Purge()

	// Retrieve the first 1024 bytes
	s.getRange("divinacommedia.txt", 0, 1024)

	// Check cache
	s.checkCacheTextFile()

	// Retrieve the text file across 2 sio packages (including the first)
	s.getRange("divinacommedia.txt", 60000, 8000)

	// Retrieve the text file across multiple sio packages, but not the first
	s.getRange("divinacommedia.txt", 100000, 100000)

	// Repeat the last test but with the metadata cache cleared
	s.cache.Purge()
	s.getRange("divinacommedia.txt", 100000, 100000)
	s.checkCacheTextFile()

	// Read till end of file
	s.getRange("divinacommedia.txt", 600100, 0)

	// Read from the first package till end
	s.getRange("divinacommedia.txt", 6000, 0)

	// Error: empty name
	_, _, err := s.store.GetWithRange(context.Background(), "", nil, nil, nil)
	if !assert.Error(s.t, err) {
		s.t.FailNow()
	}

	// File does not exist
	found, _, err := s.store.GetWithRange(context.Background(), "no_exist", nil, nil, nil)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.False(s.t, found)

	// Short file (this is less than a single package)
	s.getRange("short.txt", 0, 100)

	// Context canceled
	{
		rng := &RequestRange{
			Start:  100000,
			Length: 0,
		}
		cbCalled := false
		ctx, cancel := context.WithCancel(context.Background())
		pr, pw := io.Pipe()
		go func() {
			// Read the first 10 bytes, then cancel the context, then copy the rest
			buf := make([]byte, 10)
			io.ReadFull(pr, buf)
			cancel()
			io.Copy(ioutil.Discard, pr)
		}()
		_, _, err = s.store.GetWithRange(ctx, "kitera-dent-BIj4LObC6es-unsplash.jpg", pw, rng, func(metadata *crypto.Metadata, metadataSize int32) {
			cbCalled = (metadataSize > 0)
		})
		if !assert.Error(s.t, err) ||
			!assert.Equal(s.t, context.Canceled, err) ||
			!assert.True(s.t, cbCalled) {
			s.t.FailNow()
		}
	}
}

// Set info file
func (s *testFs) testDelete() {
	var (
		err   error
		found bool
	)
	// Delete the file
	err = s.store.Delete(context.Background(), "divinacommedia.txt", nil)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}

	// File should be deleted
	found, _, err = s.store.Get(context.Background(), "divinacommedia.txt", nil, nil)
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	if !assert.False(s.t, found) {
		s.t.FailNow()
	}

	// Error: empty name
	err = s.store.Delete(context.Background(), "", nil)
	if !assert.Error(s.t, err) {
		s.t.FailNow()
	}
}

// Used by a few tests to check that the metadata for divinacommedia.txt is present
func (s *testFs) checkCacheTextFile() {
	assert.Len(s.t, s.cache.Keys(), 1)
	headerVersion, headerLen, wrappedKey, metadataLength, metadata := s.cache.Get("divinacommedia.txt")
	assert.Equal(s.t, uint16(2), headerVersion)
	assert.Equal(s.t, int32(72), headerLen)
	assert.Len(s.t, wrappedKey, 40)
	assert.Equal(s.t, int32(38), metadataLength)
	assert.NotNil(s.t, metadata)
	assert.Equal(s.t, "divinacommedia.txt", metadata.Name)
	assert.Equal(s.t, "text/plain", metadata.ContentType)
	assert.Equal(s.t, int64(len(s.files["divinacommedia.txt"])), metadata.Size)
}

// Used by testGetWithRange
func (s *testFs) getRange(name string, start, length int64) {
	var out io.Writer = &bytes.Buffer{}
	rng := &RequestRange{
		Start:  start,
		Length: length,
	}
	cbCalled := false
	found, _, err := s.store.GetWithRange(context.Background(), name, out, rng, func(metadata *crypto.Metadata, metadataSize int32) {
		assert.Equal(s.t, name, metadata.Name)
		assert.Equal(s.t, "text/plain", metadata.ContentType)
		assert.Equal(s.t, int64(len(s.files[name])), metadata.Size)
		cbCalled = true
	})
	if !assert.NoError(s.t, err) ||
		!assert.True(s.t, found) ||
		!assert.True(s.t, cbCalled) {
		s.t.FailNow()
	}
	read, err := ioutil.ReadAll(out.(io.ReadWriter))
	if !assert.NoError(s.t, err) {
		s.t.FailNow()
	}
	assert.True(s.t, len(read) > 0)
	if length > 0 {
		assert.True(s.t, bytes.Equal(read, s.files[name][start:(start+length)]))
	} else {
		assert.True(s.t, bytes.Equal(read, s.files[name][start:]))
	}
}

// Return the info file object
func staticInfoFile() *infofile.InfoFile {
	// Create an info file with fixed data
	// Passphrase is "hello world"
	masterKey, _ := base64.StdEncoding.DecodeString("QGRFye4ebTr6U85Ja8V5d0ciZfDLXFz8gTjpqj+b6l1/N8q6oYC2hA==")
	salt, _ := base64.StdEncoding.DecodeString("Id5gT91MIeqMG7Pc1UFc8Q==")
	confirmationHash, _ := base64.StdEncoding.DecodeString("WL539+dtEvM5VDQ9LtCepF7nguCZMEzISvnFMK4UIeE=")
	return &infofile.InfoFile{
		App:      "prvt",
		Version:  4,
		DataPath: "data",
		RepoId:   "26346eac-6526-4093-a7b8-4640d4fa2f32",
		Keys: []infofile.InfoFileKey{
			{
				MasterKey:        masterKey,
				Salt:             salt,
				ConfirmationHash: confirmationHash,
			},
		},
	}
}

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandString generates a random string with the letters above
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
