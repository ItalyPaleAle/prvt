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
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// How long to cache files for, in seconds
const cacheDuration = 300

// IndexElements represents a value in the index
type IndexElement struct {
	Path string `json:"p"`
	Name string `json:"n"`
}

// CachedIndex contains the cached values for the index files
type IndexFile struct {
	Version  int            `json:"version"`
	Elements []IndexElement `json:"elements"`
}

// Index manages the index for all files and folders
type Index struct {
	cacheTime  time.Time
	cache      *IndexFile
	refreshing bool
}

// FolderList contains the result of the ListFolder method
type FolderList struct {
	Path      string `json:"path"`
	Directory bool   `json:"isDir,omitempty"`
	FileId    string `json:"fileId,omitempty"`
}

// Refresh an index if necessary
func (i *Index) Refresh(force bool) error {
	// If we're already refreshing the cache, wait
	for i.refreshing {
		time.Sleep(100 * time.Millisecond)
	}
	// Semaphore
	i.refreshing = true
	defer func() {
		i.refreshing = false
	}()

	// Check if we already have the index in cache and its age (unless we're forcing a refresh)
	if !force && i.cache != nil && time.Now().Add(cacheDuration*time.Second).Before(i.cacheTime) {
		// Cache exists and it's fresh
		return nil
	}

	// Need to request the index
	now := time.Now()
	data, err := ioutil.ReadFile("test/index")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	// Empty index
	if len(data) == 0 {
		i.cache = &IndexFile{
			Version:  1,
			Elements: []IndexElement{},
		}
		i.cacheTime = now
		return nil
	}
	i.cache = &IndexFile{}
	err = json.Unmarshal(data, i.cache)
	if err != nil {
		return err
	}
	i.cacheTime = now

	return nil
}

// Save an index object
func (i *Index) save(obj *IndexFile) error {
	now := time.Now()

	// TODO: ENCRYPT INDEX

	// Represent the data as JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// Save the updated index
	if err := ioutil.WriteFile("test/index", data, 0644); err != nil {
		return err
	}

	// Update the index in cache too
	i.cache = obj
	i.cacheTime = now

	return nil
}

// AddFile adds a file to the index
func (i *Index) AddFile(path string, fileId string) error {
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
	exists, err := i.FileExists(path)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("file already exists")
	}

	// Add the file to the index and return the id
	fileEl := IndexElement{
		Path: path,
		Name: fileId,
	}
	elements := append(i.cache.Elements, fileEl)
	updated := &IndexFile{
		Version:  1,
		Elements: elements,
	}
	if err := i.save(updated); err != nil {
		return err
	}

	return nil
}

// FileExists returns true if the file exists in the index
func (i *Index) FileExists(path string) (bool, error) {
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return false, errors.New("path must start with /")
	}

	// Iterate through the list of elemets to check if the file exists
	for _, el := range i.cache.Elements {
		// Check if there's an exact match, or if there's a folder starting with the path
		if el.Path == path || strings.HasPrefix(el.Path, path+"/") {
			return true, nil
		}
	}

	return false, nil
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

	// Force a refresh of the index
	if err := i.Refresh(false); err != nil {
		return nil, err
	}

	// Iterate through the folders looking for the one
	result := make([]FolderList, 0)
	for _, el := range i.cache.Elements {
		if strings.HasPrefix(el.Path, path) {
			// Prefix matches, so it's in the right path
			// Return only one level of sub-folders
			slashPos := strings.Index(el.Path[len(path):], "/")
			oneLevel := ""
			if slashPos == -1 {
				// No more slashes in the path
				// Means we have a file
				oneLevel = el.Path[len(path):]

				// Since we have a file, we're sure there aren't more with the same path
				result = append(result, FolderList{
					Path:      el.Path,
					Directory: false,
					FileId:    el.Name,
				})
			} else {
				// We have a directory
				// Get only until the slash
				oneLevel = el.Path[len(path):(len(path) + slashPos)]

				// Check if the path is already in the result
				if !folderListContains(result, oneLevel) {
					result = append(result, FolderList{
						Path:      oneLevel,
						Directory: true,
					})
				}
			}
		}
	}

	if len(result) == 0 {
		result = nil
	}

	return result, nil
}

// Check if a path is already contained in a []FolderList sllice
func folderListContains(list []FolderList, path string) bool {
	for _, el := range list {
		if el.Path == path {
			return true
		}
	}
	return false
}
