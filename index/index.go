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
	"crypto/sha256"
	"errors"
	"strings"
	"sync"
	"time"

	pb "github.com/ItalyPaleAle/prvt/index/proto"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Number of files in each chunk of the index
var ChunkSize = uint32(200)

// Threshold for compacting the index
// After deleting files, if the % of the deleted files is higher than this, the index will be compacted automatically
var CompactThreshold = 0.25

// FolderList contains the result of the ListFolder method
type FolderList struct {
	Path      string     `json:"path"`
	Directory bool       `json:"isDir,omitempty"`
	FileId    string     `json:"fileId,omitempty"`
	Date      *time.Time `json:"date,omitempty"`
	MimeType  string     `json:"mimeType,omitempty"`
	Digest    []byte     `json:"digest,omitempty"`
	Size      int64      `json:"size,omitempty"`
}

// IndexStats contains the result of the
type IndexStats struct {
	// Number of files in the repo
	FileCount int `json:"fileCount"`
}

// Interface for index providers, that interface with the back-end store
type IndexProvider interface {
	// Get the index file
	Get(ctx context.Context, sequence uint32) (data []byte, isJSON bool, tag interface{}, err error)
	// Set the index file
	Set(ctx context.Context, data []byte, sequence uint32, tag interface{}) (newTag interface{}, err error)
}

// Type for transaction ID
type IndexTxId int64

// Index manages the index for all files and folders
type Index struct {
	elements   []*pb.IndexElement
	deleted    []int
	cacheFiles map[string]*pb.IndexElement
	cacheTree  *IndexTreeNode
	fileHash   [][]byte
	fileTag    []interface{}
	provider   IndexProvider
	semaphore  sync.Mutex
	tx         IndexTxId
}

// SetProvider sets the providerobject to use
func (i *Index) SetProvider(provider IndexProvider) {
	// Semaphore
	i.semaphore.Lock()
	defer i.semaphore.Unlock()

	// Set the new provider object
	i.provider = provider

	// Reset the object
	i.elements = make([]*pb.IndexElement, 0)
	i.deleted = make([]int, 0)
	i.cacheFiles = nil
	i.cacheTree = nil
	i.fileHash = make([][]byte, 0)
	i.fileTag = make([]interface{}, 0)
}

// BeginTransaction starts a transaction, acquiring a lock, and returns the transaction ID
func (i *Index) BeginTransaction() IndexTxId {
	// Acquire the lock
	i.semaphore.Lock()

	// Generate a transaction ID, which is the current time
	i.tx = IndexTxId(time.Now().Unix())
	return i.tx
}

// CommitTransaction commits the changes (saving them to disk) and releases the lock
func (i *Index) CommitTransaction(tx IndexTxId) error {
	if i.tx == 0 {
		return errors.New("no transaction is currently open")
	}
	if i.tx != tx {
		return errors.New("invalid transaction id")
	}

	// Save changes
	err := i.save()

	// Release the lock at the end and delete the transaction ID
	i.tx = 0
	i.semaphore.Unlock()

	return err
}

// Refresh an index if necessary
func (i *Index) Refresh(tx IndexTxId, force bool) error {
	// Semaphore - can be skipped if we're in the transaction
	if i.tx == 0 || i.tx != tx {
		i.semaphore.Lock()
		defer i.semaphore.Unlock()
	}

	// We can force a refresh by deleting the cache first
	if force {
		i.elements = make([]*pb.IndexElement, 0)
	}

	// Do the refresh
	return i.refresh()
}

// Internal function that performs the refresh
func (i *Index) refresh() error {
	// Abort if no provider
	if i.provider == nil {
		return errors.New("provider is not initialized")
	}

	// Check if we already have the index in cache
	if len(i.elements) > 0 {
		// Cache exists and it's fresh
		return nil
	}

	// Need to request the various files in the index
	done := false
	i.elements = make([]*pb.IndexElement, 0)
	i.fileHash = make([][]byte, 0)
	i.fileTag = make([]interface{}, 0)
	for j := uint32(0); !done; j++ {
		data, isJSON, tag, err := i.provider.Get(context.Background(), j)
		if err != nil {
			return err
		}

		// Empty index
		if j == 0 && len(data) == 0 {
			i.fileHash = [][]byte{{0}}
			i.fileTag = []interface{}{nil}

			// No need to continue
			done = true
			break
		}

		// Parse a legacy JSON file or a new protobuf-encoded one
		if isJSON {
			// Only the first file can be encoded as JSON
			if j != 0 {
				return errors.New("only index file 0 can be JSON-encoded")
			}
			file := &pb.IndexFile{}
			err = protojson.Unmarshal(data, file)
			if err != nil {
				return err
			}

			// Need to iterate through all Elements and convert the Name from the UUID represented as string to bytes
			for _, el := range file.Elements {
				if el.FileIdString != "" && len(el.FileId) == 0 {
					u, err := uuid.FromString(el.FileIdString)
					if err != nil {
						return err
					}
					el.FileIdString = ""
					el.FileId = u.Bytes()
				}
			}

			// Store in cache
			i.elements = file.Elements

			// JSON-encoded indexes can not have multiple sequences, so stop here
			// No need to calculate the hash as we'll definitely need to re-encode this
			i.fileHash = [][]byte{{0}}
			i.fileTag = []interface{}{nil}
			done = true
			break
		}

		// This file is encoded as protobuf
		// Unmarshal the response
		file := &pb.IndexFile{}
		err = proto.Unmarshal(data, file)
		if err != nil {
			return err
		}

		// Sequence number must match
		if file.Sequence != j {
			return errors.New("sequence number mismatch")
		}

		// Add all elements to the cache
		i.elements = append(i.elements, file.Elements...)

		// Calculate the hash of this file and store that together with the tag
		h := sha256.Sum256(data)
		i.fileHash = append(i.fileHash, h[:])
		i.fileTag = append(i.fileTag, tag)

		// Check if there's another part to get
		done = !file.HasNext
	}

	// Build the tree
	i.buildTree()

	return nil
}

// Save the index
func (i *Index) save() error {
	fileHashLen := uint32(len(i.fileHash))
	fileTagLen := uint32(len(i.fileTag))
	elementsLen := uint32(len(i.elements))

	// Split the index into multiple chunks if needed
	chunks := elementsLen / ChunkSize
	if (elementsLen % ChunkSize) > 0 {
		chunks++
	}
	for j := uint32(0); j < chunks; j++ {
		start := (j * ChunkSize)
		end := ((j + 1) * ChunkSize)
		if end > elementsLen {
			end = elementsLen
		}
		hasNext := (j < (chunks - 1))
		obj := &pb.IndexFile{
			Version:  3,
			Sequence: j,
			HasNext:  hasNext,
			Elements: i.elements[start:end],
		}

		// Encode as a protocol buffer message
		data, err := proto.Marshal(obj)
		if err != nil {
			return err
		}

		// Check if the encoded data is any different
		newH := sha256.Sum256(data)
		if j < fileHashLen {
			curH := i.fileHash[j]
			if bytes.Equal(newH[:], curH) {
				// data hasn't changed, so move to the next chunk
				continue
			}
		}

		// Encrypt and save the updated index, if the tag is the same
		var curTag interface{}
		if j < fileTagLen {
			curTag = i.fileTag[j]
		}
		newTag, err := i.provider.Set(context.Background(), data, j, curTag)
		if err != nil {
			return err
		}

		// Update the cached data
		if j < fileHashLen {
			i.fileHash[j] = newH[:]
		} else {
			i.fileHash = append(i.fileHash, newH[:])
		}
		if j < fileTagLen {
			i.fileTag[j] = newTag
		} else {
			i.fileTag = append(i.fileTag, newTag)
		}
	}

	return nil
}

// AddFile adds a file to the index
func (i *Index) AddFile(tx IndexTxId, path string, fileId []byte, mimeType string, size int64, digest []byte, force bool) error {
	// Semaphore - can be skipped if we're in the transaction
	if i.tx == 0 || i.tx != tx {
		i.semaphore.Lock()
		defer i.semaphore.Unlock()
	}

	// path must be at least 2 characters and start with /
	if len(path) < 2 || !strings.HasPrefix(path, "/") {
		return errors.New("path must start with /")
	}
	// Ensure the path does not end with /
	if strings.HasSuffix(path, "/") {
		return errors.New("path must not end with /")
	}
	// File size must not be negative (but can be empty)
	if size < 0 {
		return errors.New("invalid file size")
	}
	// If the digest is empty, ensure it's null
	if len(digest) < 1 {
		digest = nil
	}

	// Refresh the index if needed
	if err := i.refresh(); err != nil {
		return err
	}

	// Check if the file already exists (unless we're forcing this)
	if !force {
		exists, err := i.getFileByPath(path)
		if err != nil {
			return err
		}
		// Path "/" always exists
		if exists != nil || path == "/" {
			return errors.New("file already exists")
		}
	} else if path == "/" {
		// We still can't accept a path of "/"
		return errors.New("file already exists")
	}

	// Add the file to the index
	fileEl := &pb.IndexElement{
		Path:   path,
		FileId: fileId,
		Date: &timestamppb.Timestamp{
			Seconds: time.Now().Unix(),
		},
		MimeType: mimeType,
		Size:     size,
		Digest:   digest,
	}

	// Check if we have any unused slot
	if len(i.deleted) > 0 {
		// Pop from the start of the slice
		idx := i.deleted[0]
		i.deleted = i.deleted[1:]
		// Use that slot
		i.elements[idx] = fileEl
	} else {
		// Just append at the end
		i.elements = append(i.elements, fileEl)
	}

	// Save the updated index, unless we're in a transaction
	if i.tx == 0 || i.tx != tx {
		if err := i.save(); err != nil {
			return err
		}
	}

	// Add the file to the tree and dictionary too
	i.addToTree(fileEl)

	return nil
}

// Stat returns the stats for the repo, by reading the index
// For now, this is just the number of files
func (i *Index) Stat(tx IndexTxId) (stats *IndexStats, err error) {
	// Semaphore - can be skipped if we're in the transaction
	if i.tx == 0 || i.tx != tx {
		i.semaphore.Lock()
		defer i.semaphore.Unlock()
	}

	// Refresh the index if needed
	if err := i.refresh(); err != nil {
		return nil, err
	}

	// Count the number of files
	stats = &IndexStats{
		FileCount: len(i.cacheFiles), // Use cacheFiles because it doesn't include delete files
	}
	return
}

// GetFileByPath returns the list item object for a file, searching by its path
func (i *Index) GetFileByPath(tx IndexTxId, path string) (*FolderList, error) {
	// Semaphore - can be skipped if we're in the transaction
	if i.tx == 0 || i.tx != tx {
		i.semaphore.Lock()
		defer i.semaphore.Unlock()
	}

	// Refresh the index if needed
	if err := i.refresh(); err != nil {
		return nil, err
	}

	return i.getFileByPath(path)
}

// Internal function that actually performs the lookup
func (i *Index) getFileByPath(path string) (*FolderList, error) {
	// Remove the trailing / if present
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("path must start with /")
	}

	// Iterate through the path to find the element in the tree
	// Skip the first character, which is a /
	start := 1
	node := i.cacheTree
	for y := 1; y < len(path); y++ {
		// There's a delimiter, so we have an intermediate folder
		if path[y] == '/' && (y-start) > 0 {
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
		return i.getFileById(fileId.String())
	}

	return nil, nil
}

// GetFileById returns the list item object for a file, searching by its id
func (i *Index) GetFileById(tx IndexTxId, fileId string) (*FolderList, error) {
	// Semaphore - can be skipped if we're in the transaction
	if i.tx == 0 || i.tx != tx {
		i.semaphore.Lock()
		defer i.semaphore.Unlock()
	}

	// Refresh the index if needed
	if err := i.refresh(); err != nil {
		return nil, err
	}

	return i.getFileById(fileId)
}

// Internal function that actually performs the lookup
func (i *Index) getFileById(fileId string) (*FolderList, error) {
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
		Size:      el.Size,
		Digest:    el.Digest,
	}, nil
}

// DeleteFile removes a file or folder from the index
// It returns the list of objects to remove as first argument, and their paths as second
// To remove a folder, make sure the path ends with /*
func (i *Index) DeleteFile(tx IndexTxId, path string) ([]string, []string, error) {
	// Semaphore - can be skipped if we're in the transaction
	if i.tx == 0 || i.tx != tx {
		i.semaphore.Lock()
		defer i.semaphore.Unlock()
	}

	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return nil, nil, errors.New("path must start with /")
	}

	// Refresh the index if needed
	if err := i.refresh(); err != nil {
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
	for j, el := range i.elements {
		// Need to remove
		if el.Path == path || (matchPrefix && strings.HasPrefix(el.Path, path)) {
			// Add to the result
			fileId, err := uuid.FromBytes(el.FileId)
			if err != nil {
				return nil, nil, err
			}
			objectsRemoved = append(objectsRemoved, fileId.String())
			pathsRemoved = append(pathsRemoved, el.Path)

			// Remove from the tree
			i.removeFromTree(el)

			// Add to the list of deleted files, preserving the order
			var k = 0
			for k < len(i.deleted) {
				// The == case should not happen
				if j < i.deleted[k] {
					break
				}
				k++
			}
			i.deleted = append(i.deleted, 0)
			copy(i.deleted[(k+1):], i.deleted[k:])
			i.deleted[k] = j

			// Mark the field as deleted, but do not remove the record from the list
			// In fact, doing so would cause a shift that would cause us to re-upload many more chunks than we'd need
			el.MarkDeleted()
		}
	}

	// Save if needed
	if len(objectsRemoved) > 0 {
		// Check if we need to compact the index first
		del := float64(len(i.deleted)) / float64(len(i.elements))
		if del > CompactThreshold {
			i.compact()
		}

		// Save the updated index, unless we're in a transaction
		if i.tx == 0 || i.tx != tx {
			err := i.save()
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return objectsRemoved, pathsRemoved, nil
}

// ListFolder returns the list of elements in a folder
func (i *Index) ListFolder(tx IndexTxId, path string) ([]FolderList, error) {
	// Semaphore - can be skipped if we're in the transaction
	if i.tx == 0 || i.tx != tx {
		i.semaphore.Lock()
		defer i.semaphore.Unlock()
	}

	// Ensure the path starts with a /
	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("path must start with /")
	}

	// Ensure there's a trailing slash
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Refresh the index if needed
	if err := i.refresh(); err != nil {
		return nil, err
	}

	// Iterate through the path looking for the node in the tree
	start := 1
	node := i.cacheTree
	if path != "/" {
		// Skip the first character, which is a /
		for y := 1; y < len(path); y++ {
			// Folder delimiter
			if path[y] == '/' && (y-start) > 0 {
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
	if len(node.Children) == 0 {
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
				Size:      el.File.Size,
				Digest:    el.File.Digest,
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

// Compact the index by purging all deleted records
func (i *Index) Compact(tx IndexTxId) error {
	// Semaphore - can be skipped if we're in the transaction
	if i.tx == 0 || i.tx != tx {
		i.semaphore.Lock()
		defer i.semaphore.Unlock()
	}

	// Refresh the index if needed
	if err := i.refresh(); err != nil {
		return err
	}

	// Perform the compacting
	i.compact()

	// Deleted elements were already not in the tree, so no need to rebuild that
	// Instead, just save the updated index (unless we're in a transaction)
	if i.tx == 0 || i.tx != tx {
		if err := i.save(); err != nil {
			return err
		}
	}

	return nil
}

// Builds the tree and the dictionary for easier searching
func (i *Index) buildTree() {
	// Init the objects
	i.deleted = make([]int, 0)
	i.cacheFiles = make(map[string]*pb.IndexElement, len(i.elements))
	i.cacheTree = NewIndexRootNode()

	// Iterate through the elements and build the tree
	for j, el := range i.elements {
		// Ignore deleted elements
		if !el.Deleted {
			i.addToTree(el)
		} else {
			i.deleted = append(i.deleted, j)
		}
	}
}

func (i *Index) addToTree(el *pb.IndexElement) {
	// Ensure we have a file ID and that the path begins with /
	if el.FileId == nil || len(el.Path) < 2 || el.Path[0] != '/' {
		return
	}

	// Iterate through the path to get the intermediate folders (skipping the first character which is a / itself)
	start := 1
	node := i.cacheTree
	for y := 1; y < len(el.Path); y++ {
		// There's a delimiter, so we have an intermediate folder
		if el.Path[y] == '/' && (y-start) > 0 {
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

func (i *Index) removeFromTree(el *pb.IndexElement) {
	// Ensure we have a path and that it begins with /
	if len(el.Path) == 0 || el.Path[0] != '/' {
		return
	}

	// Iterate through the path to get the intermediate folders (skipping the first character which is a / itself)
	start := 1
	stack := []*IndexTreeNode{i.cacheTree}
	node := i.cacheTree
	for y := 1; y < len(el.Path); y++ {
		// There's a delimiter, so we have an intermediate folder
		if el.Path[y] == '/' && (y-start) > 0 {
			part := el.Path[start:y]
			node = node.Find(part)
			// This should never happen, so just return
			if node == nil {
				return
			}
			stack = append(stack, node)
			start = y + 1
		}
	}

	// Whatever is left is the name of the file
	fileName := el.Path[start:]

	// Remove the leaf node
	node.Remove(fileName)

	// Go down the stack and remove all empty elements in-between
	j := len(stack)
	// Go until 1 because the first element is the root node
	for j > 1 {
		j--
		// Check if file is nil to ensure this was a folder, and then check if there are other children to see if it's empty
		// Always exclude the root node
		if stack[j].File == nil && len(stack[j].Children) == 0 {
			// Remove this empty non-root node by asking its parent
			stack[j-1].Remove(stack[j].Name)
		}
	}

	// Remove from the dictionary too
	fileIdObj, err := uuid.FromBytes(el.FileId)
	if err != nil {
		return
	}
	key := fileIdObj.String()
	delete(i.cacheFiles, key)
}

// Performs the index compacting (without locks or saving the data)
func (i *Index) compact() {
	// First, empty the deleted slice
	i.deleted = make([]int, 0)

	// Iterate through the elements and remove the deleted ones
	j := 0
	for y := 0; y < len(i.elements); y++ {
		// Keep non-deleted elements only
		if !i.elements[y].Deleted {
			i.elements[j] = i.elements[y]
			j++
		}
	}
	i.elements = i.elements[:j]
}
