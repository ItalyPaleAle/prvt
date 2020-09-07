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

package index

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// How long to cache files for, in seconds
const cacheDuration = 300

// FolderList contains the result of the ListFolder method
type FolderList struct {
	Path      string     `json:"path"`
	Directory bool       `json:"isDir,omitempty"`
	FileId    string     `json:"fileId,omitempty"`
	Date      *time.Time `json:"date,omitempty"`
	MimeType  string     `json:"mimeType,omitempty"`
}

// Index manages the index for all files and folders
type Index struct {
	cache      *IndexFile
	cacheFiles map[string]*IndexElement
	cacheTree  *IndexTreeNode
	cacheTime  time.Time
	cacheTag   interface{}
	store      fs.Fs
	semaphore  sync.Mutex
}

// SetStore sets the store (filesystem) object to use
func (i *Index) SetStore(store fs.Fs) {
	// Do not alter this if there's a refresh running
	i.semaphore.Lock()

	// Set the new store object
	i.store = store

	// Reset the cache
	i.cache = nil
	i.cacheFiles = nil
	i.cacheTree = nil
	i.cacheTag = nil

	i.semaphore.Unlock()
}

// Refresh an index if necessary
func (i *Index) Refresh(force bool) error {
	// Abort if no store
	if i.store == nil {
		return errors.New("store is not initialized")
	}

	// Semaphore
	i.semaphore.Lock()
	defer func() {
		i.semaphore.Unlock()
	}()

	// Check if we already have the index in cache and its age (unless we're forcing a refresh)
	if !force && i.cache != nil && time.Now().Add(-cacheDuration*time.Second).Before(i.cacheTime) {
		// Cache exists and it's fresh
		return nil
	}

	// Need to request the index
	now := time.Now()
	var data []byte
	buf := &bytes.Buffer{}
	isJSON := false
	found, tag, err := i.store.Get(context.Background(), "_index", buf, func(metadata *crypto.Metadata, metadataSize int32) {
		// Check if we're decoding a legacy JSON file
		if metadata.ContentType == "application/json" {
			isJSON = true
		}
	})
	if found {
		// Check error here because otherwise we might have an error also if the index wasn't found
		if err != nil {
			return err
		}

		data, err = ioutil.ReadAll(buf)
		if err != nil {
			return err
		}
	} else {
		// Ignore "not found" errors
		err = nil
	}

	// Empty index
	if len(data) == 0 {
		i.cache = &IndexFile{
			Version:  2,
			Elements: make([]*IndexElement, 0),
		}
		i.cacheTime = now
		// Build the tree
		i.buildTree()
		return nil
	}
	i.cache = &IndexFile{}

	// Parse a legacy JSON file or a new protobuf-encoded one
	if isJSON {
		err = protojson.Unmarshal(data, i.cache)
		if err != nil {
			return err
		}

		// Need to iterate through all Elements and convert the Name from the UUID represented as string to bytes
		for _, el := range i.cache.Elements {
			if el.FileIdString != "" && len(el.FileId) == 0 {
				u, err := uuid.FromString(el.FileIdString)
				if err != nil {
					return err
				}
				el.FileIdString = ""
				el.FileId = u.Bytes()
			}
		}
	} else {
		err = proto.Unmarshal(data, i.cache)
		if err != nil {
			return err
		}
	}
	i.cacheTime = now
	i.cacheTag = tag

	// Build the tree
	i.buildTree()

	return nil
}

// Save an index object
func (i *Index) save(obj *IndexFile) error {
	now := time.Now()

	// Encode the data as a protocol buffer message
	data, err := proto.Marshal(obj)
	if err != nil {
		return err
	}

	// Encrypt and save the updated index, if the tag is the same
	metadata := &crypto.Metadata{
		Name:        "index",
		ContentType: "application/protobuf",
		Size:        int64(len(data)),
	}
	buf := bytes.NewBuffer(data)
	tag, err := i.store.Set(context.Background(), "_index", buf, i.cacheTag, metadata)
	if err != nil {
		return err
	}

	// Update the index in cache too
	i.cache = obj
	i.cacheTime = now
	i.cacheTag = tag

	return nil
}

// AddFile adds a file to the index
func (i *Index) AddFile(path string, fileId []byte, mimeType string) error {
	// path must be at least 2 characters (with / being one)
	if len(path) < 2 {
		return errors.New("path name is too short")
	}
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return errors.New("path must start with /")
	}
	// Ensure the path does not end with /
	if strings.HasSuffix(path, "/") {
		return errors.New("path must not end with /")
	}

	// Force a refresh of the index
	if err := i.Refresh(true); err != nil {
		return err
	}

	// Check if the file already exists
	exists, err := i.GetFileByPath(path)
	if err != nil {
		return err
	}
	// Path "/" always exists
	if exists != nil || path == "/" {
		return errors.New("file already exists")
	}

	// Add the file to the index
	fileEl := &IndexElement{
		Path:   path,
		FileId: fileId,
		Date: &timestamppb.Timestamp{
			Seconds: time.Now().Unix(),
		},
		MimeType: mimeType,
	}
	elements := append(i.cache.Elements, fileEl)

	// Save the updated index
	updated := &IndexFile{
		Version:  2,
		Elements: elements,
	}
	if err := i.save(updated); err != nil {
		return err
	}

	// Add the file to the tree and dictionary too
	i.addToTree(fileEl)

	return nil
}

// GetFileByPath returns the list item object for a file, searching by its path
func (i *Index) GetFileByPath(path string) (*FolderList, error) {
	// Remove the trailing / if present
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("path must start with /")
	}

	// Refresh the index if needed
	if err := i.Refresh(false); err != nil {
		return nil, err
	}

	// Iterate through the path to find the element in the tree
	// Skip the first character, which is a /
	start := 1
	node := i.cacheTree
	for y := 1; y < len(path); y++ {
		// There's a delimiter, so we have an intermediate folder (and ignore double delimiters)
		if path[y] == '/' && (y-start) > 1 {
			part := path[start:y]
			if found := node.Find(part); found != nil {
				node = found
			} else {
				// Not found
				return nil, nil
			}
			start = y + 1
		}
	}

	// Last element at the end of the path
	part := path[start:]
	if found := node.Find(part); found != nil && found.File != nil && found.File.FileId != nil {
		// Get the file by its ID
		fileId, err := uuid.FromBytes(found.File.FileId)
		if err != nil {
			return nil, err
		}
		return i.GetFileById(fileId.String())
	}

	return nil, nil
}

// GetFileById returns the list item object for a file, searching by its id
func (i *Index) GetFileById(fileId string) (*FolderList, error) {
	// Refresh the index if needed
	if err := i.Refresh(false); err != nil {
		return nil, err
	}

	// Do a lookup in the dictionary
	el, found := i.cacheFiles[fileId]
	if !found || el == nil {
		return nil, nil
	}

	// Date
	var date *time.Time
	if el.Date != nil && el.Date.Seconds > 0 {
		o := time.Unix(el.Date.Seconds, 0).UTC()
		date = &o
	}

	return &FolderList{
		Path:      el.Path,
		Directory: false,
		FileId:    fileId,
		Date:      date,
		MimeType:  el.MimeType,
	}, nil
}

// DeleteFile removes a file or folder from the index
// It returns the list of objects to remove as first argument, and their paths as second
// To remove a folder, make sure the path ends with /*
func (i *Index) DeleteFile(path string) ([]string, []string, error) {
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return nil, nil, errors.New("path must start with /")
	}

	// Force a refresh of the index
	if err := i.Refresh(true); err != nil {
		return nil, nil, err
	}

	// If the path ends with /* we are going to remove the entire folder
	matchPrefix := false
	if strings.HasSuffix(path, "/*") {
		matchPrefix = true
		path = path[0 : len(path)-1]
	} else if strings.HasSuffix(path, "/") {
		return nil, nil, errors.New("USER path cannot end with /; to remove a folder, end with /*")
	} else if strings.HasSuffix(path, "*") {
		return nil, nil, errors.New("USER path cannot end with *: removing globs is supported only for folders using /* as suffix")
	}

	// Iterate through the list of files to find matches
	objectsRemoved := make([]string, 0)
	pathsRemoved := make([]string, 0)
	// Output index; see: https://stackoverflow.com/a/20551116/192024
	j := 0
	for _, el := range i.cache.Elements {
		// Need to remove
		if el.Path == path || (matchPrefix && strings.HasPrefix(el.Path, path)) {
			// Add to the result
			fileId, err := uuid.FromBytes(el.FileId)
			if err != nil {
				return nil, nil, err
			}
			objectsRemoved = append(objectsRemoved, fileId.String())
			pathsRemoved = append(pathsRemoved, el.Path)
		} else {
			// Maintain in the list
			i.cache.Elements[j] = el
			j++
		}
	}
	i.cache.Elements = i.cache.Elements[:j]

	// Save
	if err := i.save(i.cache); err != nil {
		return nil, nil, err
	}

	// Rebuild the tree and dictionary
	// TODO: Eventually, this should just update the existing tree without having to rebuild it!
	i.buildTree()

	return objectsRemoved, pathsRemoved, nil
}

// ListFolder returns the list of elements in a folder
func (i *Index) ListFolder(path string) ([]FolderList, error) {
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("path must start with /")
	}

	// Ensure there's a trailing slash
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Refresh the index if needed
	if err := i.Refresh(false); err != nil {
		return nil, err
	}

	// Iterate through the path looking for the node in the tree
	start := 1
	node := i.cacheTree
	if path != "/" {
		// Skip the first character, which is a /
		for y := 1; y < len(path); y++ {
			// Folder delimiter (and ignore double delimiters)
			if path[y] == '/' && (y-start) > 1 {
				part := path[start:y]
				if found := node.Find(part); found != nil {
					node = found
				} else {
					// Not found
					return nil, nil
				}
				start = y + 1
			}
		}
	}

	// Get the result list from the node we found
	// Note that the last character of the string is a / so we certainly have the right node
	if node == nil || node.Children == nil || len(node.Children) < 1 {
		// Nothing found
		return nil, nil
	}
	result := make([]FolderList, len(node.Children))
	y := 0
	for _, el := range node.Children {
		// We have a file
		if el.File != nil {
			// File ID
			fileId, err := uuid.FromBytes(el.File.FileId)
			if err != nil {
				return nil, err
			}

			// Date
			var date *time.Time
			if el.File.Date != nil && el.File.Date.Seconds > 0 {
				o := time.Unix(el.File.Date.Seconds, 0).UTC()
				date = &o
			}

			result[y] = FolderList{
				Path:      el.Name,
				Directory: false,
				FileId:    fileId.String(),
				Date:      date,
				MimeType:  el.File.MimeType,
			}
			y++
		} else {
			// Folder
			result[y] = FolderList{
				Path:      el.Name,
				Directory: true,
			}
			y++
		}
	}

	return result, nil
}

// Builds the tree and the dictionary for easier searching
func (i *Index) buildTree() {
	// Init the objects
	i.cacheFiles = make(map[string]*IndexElement, len(i.cache.Elements))
	i.cacheTree = &IndexTreeNode{
		Name:     "/",
		Children: make([]*IndexTreeNode, 0),
	}

	// Iterate through the elements and build the tree
	for _, el := range i.cache.Elements {
		i.addToTree(el)
	}
}

func (i *Index) addToTree(el *IndexElement) {
	// Ensure we have a file ID and that the path begins with /
	if el.FileId == nil || len(el.Path) < 2 || el.Path[0] != '/' {
		return
	}

	// Iterate through the path to get the intermediate folders (skipping the first character which is a / itself)
	start := 1
	node := i.cacheTree
	for y := 1; y < len(el.Path); y++ {
		// There's a delimiter, so we have an intermediate folder (and ignore double delimiters)
		if el.Path[y] == '/' && (y-start) > 1 {
			part := el.Path[start:y]
			if found := node.Find(part); found != nil {
				// Element exists already
				node = found
			} else {
				// Create the intermediate folder
				node = node.Add(part, nil)
			}
			start = y + 1
		}
	}

	// Whatever is left is the name of the file
	fileName := el.Path[start:]
	node.Add(fileName, el)

	// Also add to the file to the dictionary
	// The key here is the string-representation of the file ID
	fileIdObj, err := uuid.FromBytes(el.FileId)
	if err != nil {
		return
	}
	key := fileIdObj.String()
	i.cacheFiles[key] = el
}
