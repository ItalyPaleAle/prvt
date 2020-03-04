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

	"github.com/gofrs/uuid"
)

// How long to cache files for, in seconds
const cacheTime = 300

// IndexElements represents a value in the index
type IndexElement struct {
	IsFolder bool   `json:"f"`
	Name     string `json:"n"`
	Location string `json:"l"`
}

// CachedIndex contains the cached values for the index files
type CachedIndex struct {
	Time     time.Time
	Elements []IndexElement
}

// Index manages the index for all files and folders
type Index struct {
	cache      map[string]CachedIndex
	refreshing map[string]uint8
}

// Refresh an index if necessary
func (i *Index) Refresh(name string, force bool) error {
	// If we're already refreshing the cache, wait
	if i.refreshing == nil {
		i.refreshing = make(map[string]uint8)
	}
	for i.refreshing[name] != 0 {
		time.Sleep(100 * time.Millisecond)
	}
	// Semaphore
	i.refreshing[name] = 1
	defer func() {
		delete(i.refreshing, name)
	}()

	// Check if we already have the index in cache and its age (unless we're forcing a refresh)
	if i.cache == nil {
		i.cache = make(map[string]CachedIndex, 0)
	}
	if !force {
		cachedIndex, found := i.cache[name]
		if found && time.Now().Add(cacheTime*time.Second).Before(cachedIndex.Time) {
			// Cache exists and it's fresh
			return nil
		}
	}

	// Need to request the index
	var elements []IndexElement
	now := time.Now()
	data, err := ioutil.ReadFile("test/index/" + name)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	// Empty index
	if len(data) == 0 {
		// If we're requesting the list of folders, we always have at least "/"
		if name == "folders" {
			elements = []IndexElement{
				IndexElement{
					IsFolder: true,
					Name:     "/",
					Location: "root", // Convention
				},
			}
		} else {
			elements = []IndexElement{}
		}
		i.cache[name] = CachedIndex{
			Time:     now,
			Elements: elements,
		}
		return nil
	}
	err = json.Unmarshal(data, &elements)
	if err != nil {
		return err
	}
	i.cache[name] = CachedIndex{
		Time:     now,
		Elements: elements,
	}

	return nil
}

// Save an index object
func (i *Index) save(name string, elements []IndexElement) error {
	now := time.Now()

	// Represent the data as JSON
	data, err := json.Marshal(elements)
	if err != nil {
		return err
	}

	// Save the updated index
	if err := ioutil.WriteFile("test/index/"+name, data, 0644); err != nil {
		return err
	}

	// Update the index in cache too
	i.cache[name] = CachedIndex{
		Time:     now,
		Elements: elements,
	}

	return nil
}

// AddFolder adds a folder to the index
func (i *Index) AddFolder(path string) (string, error) {
	// path must not be "/"
	if path == "/" {
		return "", errors.New("cannot add root folder")
	}

	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return "", errors.New("path must start with /")
	}

	// Ensure there's a trailing slash
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Force a refresh of the folders index
	if err := i.Refresh("folders", true); err != nil {
		return "", err
	}

	// Check if the folder already exists in the index
	exists, err := i.FolderExists(path)
	if err != nil {
		return "", err
	}

	// If it already exists, return an error
	if exists {
		return "", errors.New("folder already exists")
	}

	// Ensure that all the intermediate folders (if any) exist
	parts := strings.Split(path, "/")
	for n := 0; n < len(parts)-1; n++ {
		f := "/" + strings.Join(parts[0:n], "/")
		exists, err := i.FolderExists(f)
		if err != nil {
			return "", err
		}
		if !exists {
			return "", errors.New("intermediate folder " + f + " does not exist")
		}
	}

	// Create a new folder ID for the folder, then add it to the index
	folderId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	folderIdStr := folderId.String()
	folderEl := IndexElement{
		IsFolder: true,
		Name:     path,
		Location: folderIdStr,
	}
	updated := i.cache["folders"].Elements
	updated = append(updated, folderEl)
	if err := i.save("folders", updated); err != nil {
		return "", err
	}

	return folderIdStr, nil
}

// AddFile adds a file to the index
func (i *Index) AddFile(path string) (string, error) {
	// path must be at least 2 characters (with / being one)
	if len(path) < 2 {
		return "", errors.New("path name is too short")
	}
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return "", errors.New("path must start with /")
	}
	// Ensure the path does not end with /
	if strings.HasSuffix(path, "/") {
		return "", errors.New("path must not end with /")
	}

	// Force a refresh of the folders index
	if err := i.Refresh("folders", true); err != nil {
		return "", err
	}

	// Get the file's folder and ensure it exists
	// Do not use the "FolderExists" method as we need to get the ID of the folder, to refresh the cache
	folder := Basename(path)
	var folderId string
	for _, el := range i.cache["folders"].Elements {
		// Folder exists
		if el.Name == folder && el.IsFolder {
			folderId = el.Location
			break
		}
	}
	if folderId == "" {
		return "", errors.New("folder doesn't exist")
	}

	// Force a refresh of the folder's index
	if err := i.Refresh(folderId, true); err != nil {
		return "", err
	}

	// Check if the file already exists
	exists, err := i.FileExists(path)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("file already exists")
	}

	// Add the file to the index and return the id
	fileId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	fileIdStr := fileId.String()
	fileEl := IndexElement{
		IsFolder: false,
		Name:     path,
		Location: fileIdStr,
	}
	updated := i.cache[folderId].Elements
	updated = append(updated, fileEl)
	if err := i.save(folderId, updated); err != nil {
		return "", err
	}

	return fileIdStr, nil
}

// FolderExists returns true if the folder exists in the index
func (i *Index) FolderExists(path string) (bool, error) {
	// If the folder is "/", it always exists
	if path == "/" {
		return true, nil
	}

	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return false, errors.New("path must start with /")
	}

	// Ensure there's a trailing slash
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// First, refresh the folders index, which contains the list of all folders
	if err := i.Refresh("folders", false); err != nil {
		return false, err
	}

	// Iterate throught the folders looking for the one
	for _, el := range i.cache["folders"].Elements {
		// Folder exists
		if el.Name == path && el.IsFolder {
			return true, nil
		}
	}

	return false, nil
}

// FileExists returns true if the file exists in the index
func (i *Index) FileExists(path string) (bool, error) {
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return false, errors.New("path must start with /")
	}

	// Get the folder's content
	folder := Basename(path)
	folderContents, err := i.ListFolder(folder)
	if err != nil || folderContents == nil {
		return false, err
	}

	// Check if the file exists
	for _, el := range folderContents {
		if el.Name == path {
			return true, nil
		}
	}

	return false, nil
}

// ListFolder returns the list of elements in a folder
func (i *Index) ListFolder(path string) ([]IndexElement, error) {
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("path must start with /")
	}

	// Ensure there's a trailing slash
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// First, refresh the folders index, which contains the list of all folders
	if err := i.Refresh("folders", false); err != nil {
		return nil, err
	}

	// Iterate through the folders looking for the one
	result := make([]IndexElement, 0)
	found := false
	for _, el := range i.cache["folders"].Elements {
		// Folder matches, so it has a list of files
		if el.Name == path {
			found = true
			// Load the folder's index
			if err := i.Refresh(el.Location, false); err != nil {
				return nil, err
			}
			result = append(result, i.cache[el.Location].Elements...)
		} else if strings.HasPrefix(el.Name, path) {
			// Prefix matches, so it's a sub-folder
			// Return only one level of sub-folders
			if !strings.Contains(el.Name[len(path):], "/") {
				found = true
				result = append(result, el)
			}
		}
	}

	if !found {
		result = nil
	}

	return result, nil
}
