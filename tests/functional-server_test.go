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
	"strings"
	"testing"
	"time"

	"github.com/ItalyPaleAle/prvt/cmd"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/server"

	"github.com/stretchr/testify/assert"
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
		var (
			res  *http.Response
			read []byte
		)

		// Send the request, then read the response and parse the JSON response into a map
		res, err = s.client.Get(s.serverAddr + "/api/info")
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		read, err = ioutil.ReadAll(res.Body)
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
	// Read the file
	in, err := ioutil.ReadFile(filepath.Join(s.fixtures, "short.txt"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Create the request body
	body := &bytes.Buffer{}
	mpw := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="short-text.txt"`)
	h.Set("Content-Type", "text/plain")
	partW, err := mpw.CreatePart(h)
	if err != nil {
		t.Fatal(err)
		return
	}
	_, err = partW.Write(in)
	if err != nil {
		t.Fatal(err)
		return
	}
	mpw.Close()

	// Send the request
	res, err := s.client.Post(s.serverAddr+"/api/tree/", mpw.FormDataContentType(), body)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer res.Body.Close()
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Parse the JSON response
	data := make([]server.TreeOperationReponse, 0)
	err = json.Unmarshal(read, &data)
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.Len(t, data, 1)
	assert.Equal(t, "", data[0].Error)
	assert.Equal(t, "added", data[0].Status)
	assert.Equal(t, "/short-text.txt", data[0].Path)
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
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Parse the JSON response
	data := make([]server.TreeOperationReponse, 0)
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
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Parse the JSON response
	data := make([]server.TreeOperationReponse, 0)
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
	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Parse the JSON response
	data := make([]server.TreeOperationReponse, 0)
	err = json.Unmarshal(read, &data)
	if err != nil {
		t.Fatal(err)
		return
	}

	assert.Len(t, data, 1)
	assert.Equal(t, "", data[0].Error)
	assert.Equal(t, "existing", data[0].Status)
	assert.Equal(t, "/upload/joshua-woroniecki-dyEaBD5uiio-unsplash.jpg", data[0].Path)
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
	deleteRequest := func(path string) (data []server.TreeOperationReponse, err error) {
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
		read, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		data = make([]server.TreeOperationReponse, 0)
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
		deleted       []server.TreeOperationReponse
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
