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

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ItalyPaleAle/prvt/cmd"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/server"
	"github.com/ItalyPaleAle/prvt/utils"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

// RunServer runs the sequence of tests for the server; must be run run after the CLI tests
func (s *funcTestSuite) RunServer(t *testing.T) {
	var close func()
	// Test on existing repo
	s.promptPwd.SetPasswords("hello world")
	close = s.startServer(t, "--store", "local:"+s.dirs[0])
	t.Run("info", s.serverInfo)
	t.Run("add files", s.serverAdd)
	t.Run("list and remove files", s.serverListRemove)
	t.Run("get file metadata", s.serverFileMetadata)
	t.Run("get file", s.serverFile)
	t.Run("file HEAD request", s.serverFileHeadRequest)
	t.Run("get file chunks", s.serverFileChunks)
	t.Run("interrupt getting files", s.serverFileInterrupt)
	t.Run("serving web UI", s.serverWebUI)
	close()
}

func (s *funcTestSuite) startServer(t *testing.T, args ...string) func() {
	// Start the server by invoking the command
	serveCmd := cmd.NewServeCmd()
	serveCmd.SetOut(ioutil.Discard)
	serveCmd.SetErr(ioutil.Discard)
	serveCmd.SetArgs(args)

	// Server address
	address, err := serveCmd.Flags().GetString("address")
	if err != nil {
		t.Fatal(err)
		return nil
	}
	port, err := serveCmd.Flags().GetString("port")
	if err != nil {
		t.Fatal(err)
		return nil
	}
	s.serverAddr = fmt.Sprintf("http://%s:%s", address, port)

	// Set a context that can be canceled, used to stop the server
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := serveCmd.ExecuteContext(ctx)
		if err != nil {
			log.Fatal("cannot start server\nargs:", args, "\nerror:", err)
			return
		}
	}()

	// Wait a couple of seconds to ensure the server has started
	time.Sleep(2 * time.Second)

	// The caller can stop the server
	return cancel
}

// Test the API info endpoint
func (s *funcTestSuite) serverInfo(t *testing.T) {
	sendRequest := func() (data map[string]string, err error) {
		// Send the request, then read the response and parse the JSON response into a map
		res, err := s.client.Get(s.serverAddr + "/api/info")
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode > 299 {
			return nil, fmt.Errorf("invalid response status code: %d", res.StatusCode)
		}
		read, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		data = make(map[string]string)
		err = json.Unmarshal(read, &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	// Check the response
	data, err := sendRequest()
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, data, 2)
	assert.Equal(t, "prvt", data["name"])
	assert.True(t, len(data["info"]) > 0)

	// Set buildinfo then check again
	reset := setBuildInfo()
	data, err = sendRequest()
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, data, 6)
	assert.Equal(t, "prvt", data["name"])
	assert.True(t, len(data["version"]) > 0)
	assert.True(t, len(data["buildId"]) > 0)
	assert.True(t, len(data["buildTime"]) > 0)
	assert.True(t, len(data["commitHash"]) > 0)
	assert.True(t, len(data["runtime"]) > 0)
	reset()
}

// Test the endpoint for adding files
func (s *funcTestSuite) serverAdd(t *testing.T) {
	t.Run("upload single file", s.serverAddUploadFile)
	t.Run("upload multiple files", s.serverAddUploadMultiFiles)
	t.Run("add local files", s.serverAddLocalFiles)
	t.Run("add one existing local file", s.serverAddLocalFileExisting)
}

// Add a file by uploading it directly, to the / folder
func (s *funcTestSuite) serverAddUploadFile(t *testing.T) {
	// Load the test file
	content, err := ioutil.ReadFile(filepath.Join(s.fixtures, "short.txt"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Upload the test file
	s.uploadFile(t, content, "short-text.txt", "/", "text/plain")
	if t.Failed() {
		t.FailNow()
		return
	}
}

// Add multiple files via direct upload, to the /upload folder
func (s *funcTestSuite) serverAddUploadMultiFiles(t *testing.T) {
	// Open the files
	var err error
	paths := []string{
		"joshua-woroniecki-dyEaBD5uiio-unsplash.jpg",
		"partha-narasimhan-kT5Syi2Ll3w-unsplash.jpg",
		"leigh-williams-CCABYukxjHs-unsplash.jpg",
	}
	files := make([]*os.File, len(paths))
	for i, p := range paths {
		files[i], err = os.Open(filepath.Join(s.fixtures, "photos", p))
		if err != nil {
			t.Fatal(err)
			return
		}
		defer files[i].Close()
	}

	// Create the request body
	pr, pw := io.Pipe()
	mpw := multipart.NewWriter(pw)
	go func() {
		var partW io.Writer
		for i := 0; i < len(paths); i++ {
			f := files[i]
			p := paths[i]
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="file"; filename="`+p+`"`)
			h.Set("Content-Type", "image/jpeg")
			partW, err = mpw.CreatePart(h)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(partW, f)
			if err != nil {
				panic(err)
			}
		}
		mpw.Close()
		pw.Close()
	}()

	// Send the request
	res, err := s.client.Post(s.serverAddr+"/api/tree/upload", mpw.FormDataContentType(), pr)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Parse the JSON response
	data := make([]server.TreeOperationResponse, 0)
	err = json.Unmarshal(read, &data)
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.Len(t, data, len(paths))
	for i, p := range paths {
		assert.Equal(t, "", data[i].Error)
		assert.Equal(t, "added", data[i].Status)
		assert.Equal(t, "/upload/"+p, data[i].Path)
		assert.True(t, data[i].FileId != "")
	}
}

// Add multiple files from the local file system, to the /added folder
func (s *funcTestSuite) serverAddLocalFiles(t *testing.T) {
	// Create the request body
	body := &bytes.Buffer{}
	mpw := multipart.NewWriter(body)
	err := mpw.WriteField("localpath", filepath.Join(s.fixtures, "photos"))
	if err != nil {
		t.Fatal(err)
		return
	}
	mpw.Close()

	// Send the request
	res, err := s.client.Post(s.serverAddr+"/api/tree/added", mpw.FormDataContentType(), body)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Parse the JSON response
	data := make([]server.TreeOperationResponse, 0)
	err = json.Unmarshal(read, &data)
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.Len(t, data, 4)
	for i := 0; i < 4; i++ {
		assert.Equal(t, "", data[i].Error)
		assert.Equal(t, "added", data[i].Status)
		assert.True(t, strings.HasPrefix(data[i].Path, "/added/photos/"))
		assert.True(t, strings.HasSuffix(data[i].Path, "-unsplash.jpg"))
		assert.True(t, data[i].FileId != "")
	}
}

// Add a single file from the local file system, to the /upload folder, already existing
func (s *funcTestSuite) serverAddLocalFileExisting(t *testing.T) {
	// Create the request body
	body := &bytes.Buffer{}
	mpw := multipart.NewWriter(body)
	err := mpw.WriteField("localpath", filepath.Join(s.fixtures, "photos", "joshua-woroniecki-dyEaBD5uiio-unsplash.jpg"))
	if err != nil {
		t.Fatal(err)
		return
	}
	mpw.Close()

	// Send the request
	res, err := s.client.Post(s.serverAddr+"/api/tree/upload/", mpw.FormDataContentType(), body)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Parse the JSON response
	data := make([]server.TreeOperationResponse, 0)
	err = json.Unmarshal(read, &data)
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.Len(t, data, 1)
	assert.Equal(t, "", data[0].Error)
	assert.Equal(t, "existing", data[0].Status)
	assert.Equal(t, "/upload/joshua-woroniecki-dyEaBD5uiio-unsplash.jpg", data[0].Path)
	assert.Equal(t, "", data[0].FileId)
}

// Test the endpoint that lists files
func (s *funcTestSuite) serverListRemove(t *testing.T) {
	listRequest := func(path string) (data []index.FolderList, err error) {
		// Send the request, then read the response and parse the JSON response
		res, err := s.client.Get(s.serverAddr + "/api/tree" + path)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode > 299 {
			return nil, fmt.Errorf("invalid response status code: %d", res.StatusCode)
		}
		read, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		data = make([]index.FolderList, 0)
		err = json.Unmarshal(read, &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	deleteRequest := func(path string) (data []server.TreeOperationResponse, err error) {
		// Send the request, then read the response and parse the JSON response
		req, err := http.NewRequest("DELETE", s.serverAddr+"/api/tree"+path, nil)
		if err != nil {
			return nil, err
		}
		res, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode > 299 {
			return nil, fmt.Errorf("invalid response status code: %d", res.StatusCode)
		}
		read, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		data = make([]server.TreeOperationResponse, 0)
		err = json.Unmarshal(read, &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	var (
		err           error
		expect, found []string
		list          []index.FolderList
		deleted       []server.TreeOperationResponse
	)

	// Request the / path
	list, err = listRequest("/")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 4)
	expect = []string{"added/", "short-text.txt", "text/", "upload/"}
	found = []string{}
	for _, e := range list {
		path := e.Path
		if e.Directory {
			path += "/"
		} else {
			assert.True(t, e.Date != nil)
			assert.True(t, e.FileId != "")
			assert.True(t, e.MimeType == "text/plain")
		}
		found = append(found, path)
	}
	sort.Strings(found)
	assert.True(t, reflect.DeepEqual(expect, found))

	// Repeat the test but without the "/" at the end of the URL
	list, err = listRequest("")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 4)
	expect = []string{"added/", "short-text.txt", "text/", "upload/"}
	found = []string{}
	for _, e := range list {
		path := e.Path
		if e.Directory {
			path += "/"
		} else {
			assert.True(t, e.Date != nil)
			assert.True(t, e.FileId != "")
			assert.True(t, e.MimeType == "text/plain")
		}
		found = append(found, path)
	}
	sort.Strings(found)
	assert.True(t, reflect.DeepEqual(expect, found))

	// List the upload folder
	list, err = listRequest("/upload")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 3)
	expect = []string{
		"joshua-woroniecki-dyEaBD5uiio-unsplash.jpg",
		"leigh-williams-CCABYukxjHs-unsplash.jpg",
		"partha-narasimhan-kT5Syi2Ll3w-unsplash.jpg",
	}
	found = []string{}
	for _, e := range list {
		assert.True(t, !e.Directory)
		assert.True(t, e.Date != nil)
		assert.True(t, e.FileId != "")
		assert.True(t, e.MimeType == "image/jpeg")
		found = append(found, e.Path)
		s.fileIds["/upload/"+e.Path] = e.FileId
	}
	sort.Strings(found)
	assert.True(t, reflect.DeepEqual(expect, found))

	// List a path that doesn't exist
	list, err = listRequest("/not-found")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 0)

	// Delete a file
	deleted, err = deleteRequest("/short-text.txt")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, deleted, 1)
	assert.Equal(t, "/short-text.txt", deleted[0].Path)
	assert.True(t, deleted[0].FileId != "")
	assert.Equal(t, "removed", deleted[0].Status)
	assert.Equal(t, "", deleted[0].Error)

	// Error: delete a file that doesn't exist
	deleted, err = deleteRequest("/not-found.txt")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, deleted, 1)
	assert.Equal(t, "/not-found.txt", deleted[0].Path)
	assert.Equal(t, "not-found", deleted[0].Status)
	assert.Equal(t, "", deleted[0].Error)

	// Error: cannot delete files with a glob
	deleted, err = deleteRequest("/short*")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, deleted, 1)
	assert.Equal(t, "/short*", deleted[0].Path)
	assert.Equal(t, "", deleted[0].FileId)
	assert.Equal(t, "error", deleted[0].Status)
	assert.True(t, strings.HasPrefix(deleted[0].Error, "Error while removing path from index: path cannot end with *"))

	// Error: to delete a folder, must end with /*
	deleted, err = deleteRequest("/added/")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, deleted, 1)
	assert.Equal(t, "/added/", deleted[0].Path)
	assert.Equal(t, "", deleted[0].FileId)
	assert.Equal(t, "error", deleted[0].Status)
	assert.True(t, strings.HasPrefix(deleted[0].Error, "Error while removing path from index: path cannot end with /"))

	// Delete an entire folder
	deleted, err = deleteRequest("/added/*")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, deleted, 4)
	for i := 0; i < 4; i++ {
		assert.Equal(t, "", deleted[i].Error)
		assert.True(t, deleted[i].FileId != "")
		assert.Equal(t, "removed", deleted[i].Status)
		assert.True(t, strings.HasPrefix(deleted[i].Path, "/added/photos/"))
		assert.True(t, strings.HasSuffix(deleted[i].Path, "-unsplash.jpg"))
	}

	// Request the / path again
	list, err = listRequest("/")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Len(t, list, 2)
	expect = []string{"text/", "upload/"}
	found = []string{}
	for _, e := range list {
		assert.True(t, e.Directory)
		found = append(found, e.Path+"/")
	}
	sort.Strings(found)
	assert.True(t, reflect.DeepEqual(expect, found))
}

// Test the file metadata endpoint
func (s *funcTestSuite) serverFileMetadata(t *testing.T) {
	sendRequest := func(file string) (data *server.MetadataResponse, err error) {
		// Send the request, then read the response and parse the JSON response into a map
		res, err := s.client.Get(s.serverAddr + "/api/metadata/" + file)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode > 299 {
			return nil, fmt.Errorf("invalid response status code: %d", res.StatusCode)
		}
		read, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		data = &server.MetadataResponse{}
		err = json.Unmarshal(read, data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	var (
		data       *server.MetadataResponse
		dataRepeat *server.MetadataResponse
		err        error
	)

	// Request metadata using a file path
	data, err = sendRequest("upload/leigh-williams-CCABYukxjHs-unsplash.jpg")
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Equal(t, "leigh-williams-CCABYukxjHs-unsplash.jpg", data.Name)
	assert.Equal(t, "/upload/", data.Folder)
	assert.Equal(t, "image/jpeg", data.MimeType)
	assert.Equal(t, int64(350990), data.Size)
	assert.True(t, data.Date != nil)
	assert.True(t, data.FileId != "")

	// Request the same but using the file id this time
	dataRepeat, err = sendRequest(data.FileId)
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.True(t, reflect.DeepEqual(data, dataRepeat))

	// Error: not found
	_, err = sendRequest("notfound")
	assert.EqualError(t, err, "invalid response status code: 404")

	// Error: empty file name
	_, err = sendRequest("")
	assert.EqualError(t, err, "invalid response status code: 400")
}

// Test retrieving whole files
func (s *funcTestSuite) serverFile(t *testing.T) {
	// Load the test file
	content, err := ioutil.ReadFile(filepath.Join(s.fixtures, "divinacommedia.txt"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Upload the test file
	fileId := s.uploadFile(t, content, "text1.txt", "/serve-test/", "text/plain")
	if t.Failed() {
		t.FailNow()
		return
	}

	// Store the file ID
	s.fileIds["/serve-test/text1.txt"] = fileId

	// Retrieve the file in full
	res, err := s.client.Get(s.serverAddr + "/file/" + fileId)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Ensure that the data retrieved is the same
	assert.Equal(t, content, read)
}

// Test HEAD requests to the file endpoint
func (s *funcTestSuite) serverFileHeadRequest(t *testing.T) {
	// Load the test file
	content, err := ioutil.ReadFile(filepath.Join(s.fixtures, "divinacommedia.txt"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Upload the test file again and store the file id
	s.fileIds["/serve-test/text2.txt"] = s.uploadFile(t, content, "text2.txt", "/serve-test/", "text/plain")
	if t.Failed() {
		t.FailNow()
		return
	}

	// Make a head request to the first file (whose metadata should be cached)
	res, err := s.client.Head(s.serverAddr + "/file/" + s.fileIds["/serve-test/text1.txt"])
	if err != nil {
		t.Fatal(err)
		return
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}

	// Check the required response headers
	assert.Equal(t, "bytes", res.Header.Get("Accept-Ranges"))
	assert.Contains(t, res.Header.Get("Content-Disposition"), `filename="text1.txt"`)
	assert.Equal(t, strconv.Itoa(len(content)), res.Header.Get("Content-Length"))
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))

	// Repeat for the second file, whose metadata isn't cached
	res, err = s.client.Head(s.serverAddr + "/file/" + s.fileIds["/serve-test/text2.txt"])
	if err != nil {
		t.Fatal(err)
		return
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}

	// Check the required response headers
	assert.Equal(t, "bytes", res.Header.Get("Accept-Ranges"))
	assert.Contains(t, res.Header.Get("Content-Disposition"), `filename="text2.txt"`)
	assert.Equal(t, strconv.Itoa(len(content)), res.Header.Get("Content-Length"))
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
}

// Test retrieving chunks
func (s *funcTestSuite) serverFileChunks(t *testing.T) {
	var (
		content, read []byte
		err           error
		req           *http.Request
		res           *http.Response
	)
	// Load the test file
	content, err = ioutil.ReadFile(filepath.Join(s.fixtures, "divinacommedia.txt"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Upload the test file again and store the file ID
	s.fileIds["/serve-test/text3.txt"] = s.uploadFile(t, content, "text3.txt", "/serve-test/", "text/plain")
	if t.Failed() {
		t.FailNow()
		return
	}

	// Retrieve a chunk of the file (from a file with cached metadata)
	// Then ensure that the data retrieved is the same
	req, err = http.NewRequest("GET", s.serverAddr+"/file/"+s.fileIds["/serve-test/text2.txt"], nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	req.Header.Add("Range", "bytes=65409-65485")
	res, err = s.client.Do(req)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Equal(t, content[65409:65486], read)

	// Retrieve a chunk of a file whose metadata wasn't cached
	// Then ensure that the data retrieved is the same
	req, err = http.NewRequest("GET", s.serverAddr+"/file/"+s.fileIds["/serve-test/text3.txt"], nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	req.Header.Add("Range", "bytes=600010-")
	res, err = s.client.Do(req)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Equal(t, content[600010:], read)
}

// Test interrupting file retrieval
func (s *funcTestSuite) serverFileInterrupt(t *testing.T) {
	leakOpts := goleak.IgnoreCurrent()
	defer func() {
		// There is a timeout to detect connections that are idle, so keep waiting until that before failing a test
		i := 0
		err := goleak.Find()
		for err != nil && i < (server.IdleTimeout*2) {
			i++
			time.Sleep(500 * time.Millisecond)
			err = goleak.Find(leakOpts)
		}
		// Check if we still have an error
		if err != nil {
			t.Fatal(err)
		}
	}()

	makeRequest := func(url string, addHeaders map[string]string) (err error) {
		// Context with a timeout
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create the request
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}
		if addHeaders != nil && len(addHeaders) > 0 {
			for k, v := range addHeaders {
				req.Header.Add(k, v)
			}
		}

		// Submit the request
		res, err := s.client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode > 299 {
			return fmt.Errorf("invalid response status code: %d", res.StatusCode)
		}

		// Read the first 1000 bytes only then cancel the request
		buf := make([]byte, 1000)
		n, err := res.Body.Read(buf)
		if err != nil {
			return err
		}
		assert.Equal(t, 1000, n)
		cancel()

		// Read the rest: should end prematurely (definitely less than 128KB) with "context canceled"
		m, err := io.Copy(ioutil.Discard, res.Body)
		assert.EqualError(t, err, "context canceled")
		assert.True(t, m < 128*1024)

		return nil
	}

	var err error

	// Retrieve the file in full
	err = makeRequest(
		s.serverAddr+"/file/"+s.fileIds["/upload/leigh-williams-CCABYukxjHs-unsplash.jpg"],
		nil,
	)
	if err != nil {
		t.Error(err)
		return
	}

	// Retrieve a chunk only
	err = makeRequest(
		s.serverAddr+"/file/"+s.fileIds["/upload/partha-narasimhan-kT5Syi2Ll3w-unsplash.jpg"],
		map[string]string{
			"Range": "bytes=70000-",
		},
	)
	if err != nil {
		t.Error(err)
		return
	}
}

// Check that the server is returning the web UI
func (s *funcTestSuite) serverWebUI(t *testing.T) {
	// Skip the test if the web UI wasn't compiled
	path := filepath.Join(s.fixtures, "..", "..", "ui", "dist", "index.html")
	exists, err := utils.PathExists(path)
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Skip("web UI not compiled")
		return
	}

	// Read the index file from disk
	read, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}

	// Request the index file from the server
	res, err := s.client.Get(s.serverAddr)
	if err != nil {
		t.Error(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Errorf("invalid response status code: %d", res.StatusCode)
		return
	}
	received, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}

	// Compare the response
	assert.Equal(t, read, received)
}

// Internal function used to upload individual files
func (s *funcTestSuite) uploadFile(t *testing.T, content []byte, filename string, dest string, contentType string) (fileId string) {
	t.Helper()

	// Create the request body
	body := &bytes.Buffer{}
	mpw := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", contentType)
	partW, err := mpw.CreatePart(h)
	if err != nil {
		t.Fatal(err)
		return ""
	}
	_, err = partW.Write(content)
	if err != nil {
		t.Fatal(err)
		return
	}
	mpw.Close()

	// Send the request
	res, err := s.client.Post(s.serverAddr+"/api/tree"+dest, mpw.FormDataContentType(), body)
	if err != nil {
		t.Fatal(err)
		return ""
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		t.Fatalf("invalid response status code: %d", res.StatusCode)
		return
	}
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return ""
	}

	// Parse the JSON response
	data := make([]server.TreeOperationResponse, 0)
	err = json.Unmarshal(read, &data)
	if err != nil {
		t.Fatal(err)
		return ""
	}

	assert.Len(t, data, 1)
	assert.Equal(t, "", data[0].Error)
	assert.Equal(t, "added", data[0].Status)
	assert.Equal(t, dest+filename, data[0].Path)
	assert.True(t, len(data[0].FileId) > 0)

	return data[0].FileId
}
