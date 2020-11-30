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
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	pb "github.com/ItalyPaleAle/prvt/index/proto"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

var i *Index
var provider *testIndexProvider
var expect []*pb.IndexElement
var fileNum = 0

// Base UUID format
const baseUUID = "%08d-0000-49ba-0000-000000000000"

// Tests for the Index class
func TestIndex(t *testing.T) {
	// Set ChunkSize to a very small value for testing
	ChunkSize = 5

	// Increase the CompactThreshold because we are working with not a lot of entries
	// Setting this to 1 effectively disables it
	CompactThreshold = 1

	// Init the objects
	i = &Index{}
	provider = &testIndexProvider{}
	provider.Init()
	i.SetProvider(provider)
	expect = make([]*pb.IndexElement, 0)

	// Ensure the object is empty
	assert.Empty(t, i.elements)
	assert.Empty(t, i.deleted)
	assert.Empty(t, i.cacheFiles)
	assert.Empty(t, i.cacheTree)
	assert.Empty(t, i.fileHash)
	assert.Empty(t, i.fileTag)

	// Add a file
	provider.changeTrackStart()
	addFile(t, "/file1.txt")
	checkIndex(t)
	checkSaved(t, []int{0})

	// Add more files (still in a single chunk)
	provider.changeTrackStart()
	addFile(t, "/file2.txt")
	addFile(t, "/sub/file.txt")
	addFile(t, "/folder/file.txt")
	addFile(t, "/sub/sub/file.txt")
	checkIndex(t)
	checkSaved(t, []int{0})

	// Adding more files; this should create another chunk
	// Chunk 0 should be changed too because it is not the last anymore
	provider.changeTrackStart()
	addFile(t, "/foo/bar/foo.txt")
	checkIndex(t)
	checkSaved(t, []int{0, 1})

	// Adding more files again, should only affect chunk 1
	provider.changeTrackStart()
	addFile(t, "/foo/bar/foo2.txt")
	addFile(t, "/folder/hello.txt")
	addFile(t, "/hello/hello-world.txt")
	addFile(t, "/hello/ciao-mondo.txt")
	checkIndex(t)
	checkSaved(t, []int{1})

	// List files in folders
	listFiles(t, "/hello/")
	listFiles(t, "/folder")
	listFiles(t, "/")
	listFiles(t, "/not-found")

	// Delete a file from the first chunk
	provider.changeTrackStart()
	deleteFile(t, "/sub/sub/file.txt")
	checkIndex(t)
	checkSaved(t, []int{0})

	// Delete two files from the second chunk
	provider.changeTrackStart()
	deleteFile(t, "/foo/bar/*")
	checkIndex(t)
	checkSaved(t, []int{1})

	// Delete files in both the first and second chunk
	provider.changeTrackStart()
	deleteFile(t, "/folder/*")
	checkIndex(t)
	checkSaved(t, []int{0, 1})

	// Add more files which should use the slots of deleted ones
	provider.changeTrackStart()
	addFile(t, "/added/1.txt")
	addFile(t, "/added/2.txt")
	addFile(t, "/added/3.txt")
	checkIndex(t)
	checkSaved(t, []int{0, 1})

	// Get a file by its ID
	el, err := i.GetFileById("00000009-0000-49ba-0000-000000000000")
	assert.NoError(t, err)
	assert.NotEmpty(t, el)
	assert.Equal(t, "/hello/hello-world.txt", el.Path)

	// Get a file by its path
	el, err = i.GetFileByPath("/hello/hello-world.txt")
	assert.NoError(t, err)
	assert.NotEmpty(t, el)
	assert.Equal(t, "/hello/hello-world.txt", el.Path)

	// Add more files; these will use the slot of deleted ones, plus more
	provider.changeTrackStart()
	addFile(t, "/added/4.txt")
	addFile(t, "/added/5.txt")
	addFile(t, "/added/6.txt")
	checkIndex(t)
	checkSaved(t, []int{1, 2})

	// List again
	listFiles(t, "/added/")

	// Set CompactThreshold to a lower number for testing it
	CompactThreshold = 0.2

	// Delete a large enough number of files to trigger compacting
	// Because we'll have only 5 items at the end, only chunk 0 will be saved
	provider.changeTrackStart()
	deleteFile(t, "/added/*")
	compactTestExpect(t)
	checkIndex(t)
	checkSaved(t, []int{0})
}

// Helper function that adds a file
func addFile(t *testing.T, path string) {
	t.Helper()

	// Add the file
	fileNum++
	fileId := fmt.Sprintf(baseUUID, fileNum)
	u, err := uuid.FromString(fileId)
	assert.NoError(t, err)
	err = i.AddFile(path, u.Bytes(), "text/plain", 100, nil, false)
	assert.NoError(t, err)
	add := &pb.IndexElement{
		// Only store the file ID and path (and the Deleted flag)
		FileId:  u.Bytes(),
		Path:    path,
		Deleted: false,
	}

	// If we have a spot from a deleted file, use that
	added := false
	for j := 0; j < len(expect); j++ {
		if expect[j].Deleted {
			expect[j] = add
			added = true
			break
		}
	}
	if !added {
		expect = append(expect, add)
	}
}

// Helper function that deletes a file
func deleteFile(t *testing.T, path string) {
	t.Helper()

	// Delete the file from the index
	objectsRemoved, _, err := i.DeleteFile(path)
	assert.NoError(t, err)

	// Remove from the expected list
	removed := 0
	for _, obj := range objectsRemoved {
		// Convert the UUID to a byte slice
		u, err := uuid.FromString(obj)
		assert.NoError(t, err)

		// Remove
		for j := 0; j < len(expect); j++ {
			if bytes.Equal(u[:], expect[j].FileId) {
				expect[j].Deleted = true
				expect[j].FileId = nil
				expect[j].Path = ""
				removed++
				break
			}
		}
	}
	assert.Equal(t, len(objectsRemoved), removed)
}

// Helper function that lists files in a folder
func listFiles(t *testing.T, path string) {
	t.Helper()

	// Get the list
	list, err := i.ListFolder(path)
	assert.NoError(t, err)

	// Get the list of expected results
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	exp := make([]string, 0)
	for j := 0; j < len(expect); j++ {
		if expect[j].Deleted {
			continue
		}
		if strings.HasPrefix(expect[j].Path, path) {
			e := expect[j].Path[len(path):]
			// Only get up to the slash
			pos := strings.IndexRune(e, '/')
			if pos > 0 {
				e = e[:pos]
			}
			if !StringInSlice(exp, e) {
				exp = append(exp, e)
			}
		}
	}

	// Get the list of paths found
	res := make([]string, len(list))
	for j, el := range list {
		res[j] = el.Path
	}

	// Sort the slices
	sort.Strings(exp)
	sort.Strings(res)

	assert.True(t, reflect.DeepEqual(exp, res))
}

// Helper function that checks the index based on what we expect
func checkIndex(t *testing.T) {
	t.Helper()

	defer func() {
		// Dump the state if the test failed
		if t.Failed() {
			i.DumpState()
		}
	}()

	// Check the elements and deleted slices and the cachedFiles map
	if len(i.elements) != len(expect) {
		t.Error("number of elements doesn't match expected")
		return
	}
	d := 0
	for j := 0; j < len(expect); j++ {
		assert.Equal(t, expect[j].Deleted, i.elements[j].Deleted)

		if !expect[j].Deleted {
			// File ID as string
			u, err := uuid.FromBytes(expect[j].FileId)
			assert.NoError(t, err)
			fileId := u.String()

			// Elements slice
			assert.NotEmpty(t, i.elements[j])
			assert.Equal(t, expect[j].Path, i.elements[j].Path)
			assert.True(t, bytes.Equal(expect[j].FileId, i.elements[j].FileId))

			// cacheFiles map
			cached, found := i.cacheFiles[fileId]
			assert.True(t, found)
			assert.Equal(t, cached, i.elements[j])
		} else {
			// If this is deleted, check the deleted slice
			if assert.Less(t, d, len(i.deleted)) {
				assert.Equal(t, j, i.deleted[d])
			}
			d++
		}
	}

	// Check stats
	existing := len(expect) - d // Number of non-deleted files
	stat, err := i.Stat()
	assert.NoError(t, err)
	assert.NotNil(t, stat)
	assert.Equal(t, existing, stat.FileCount)

	// Check stored data and each sequence
	chunks := uint32(len(expect) / int(ChunkSize))
	if (len(expect) % int(ChunkSize)) != 0 {
		chunks++
	}
	x := 0
	for j := uint32(0); j < chunks; j++ {
		res, err := provider.getSequenceContents(j)
		assert.NoError(t, err)
		assert.Equal(t, j, res.Sequence)
		assert.EqualValues(t, 3, res.Version)

		// Last chunk
		if (j + 1) == chunks {
			assert.False(t, res.HasNext)
		} else {
			assert.True(t, res.HasNext)
		}

		// Elements
		for y := 0; y < len(res.Elements); y++ {
			k := int(j*ChunkSize) + y
			assert.NotEmpty(t, res.Elements[y])
			assert.Equal(t, expect[k].Deleted, res.Elements[y].Deleted)
			assert.Equal(t, expect[k].Path, res.Elements[y].Path)
			assert.True(t, bytes.Equal(expect[k].FileId, res.Elements[y].FileId))
			x++
		}
	}
	// Must have iterated through all
	assert.Equal(t, len(expect), x)
}

// Ensures that the right chunks were saved only
func checkSaved(t *testing.T, expect []int) {
	t.Helper()

	saved := provider.changeTrackStop()

	// Sort
	sort.Ints(expect)
	sort.Ints(saved)

	// Check
	if !assert.True(t, reflect.DeepEqual(expect, saved)) {
		fmt.Println(expect, saved)
	}
}

// Purges deleted elements from the tests' state
func compactTestExpect(t *testing.T) {
	t.Helper()

	z := 0
	for j := 0; j < len(expect); j++ {
		if !expect[j].Deleted {
			expect[z] = expect[j]
			z++
		}
	}
	expect = expect[:z]
}

/* IndexProvider for the test */

type testIndexProvider struct {
	files   map[uint32][]byte
	tags    map[uint32]int
	changed map[uint32]bool
}

func (p *testIndexProvider) Init() {
	p.files = make(map[uint32][]byte)
	p.tags = make(map[uint32]int)

	// Seed the PRNG
	rand.Seed(time.Now().UnixNano())
}

func (p *testIndexProvider) Get(ctx context.Context, sequence uint32) (data []byte, isJSON bool, tag interface{}, err error) {
	// Check if we have any data
	data = p.files[sequence]
	if len(data) == 0 {
		return nil, false, nil, nil
	}

	// Retrieve the tag
	n := p.tags[sequence]
	tag = &n
	return data, false, tag, nil
}

func (p *testIndexProvider) Set(ctx context.Context, data []byte, sequence uint32, tag interface{}) (newTag interface{}, err error) {
	// Ensure data is not empty
	if len(data) == 0 {
		return nil, errors.New("data must not be empty")
	}

	// If there's a tag being passed, ensure it matches the current one
	if tag != nil {
		n, ok := tag.(*int)
		if ok {
			if p.tags[sequence] != *n {
				return nil, errors.New("tag mismatch")
			}
		}
	}

	// Store the file
	p.files[sequence] = data

	// Random tag
	rnd := rand.Int()
	newTag = &rnd
	p.tags[sequence] = rnd

	// Set that this chunk was saved
	if p.changed != nil {
		p.changed[sequence] = true
	}

	return newTag, nil
}

// Starts recording which chunks are saved
func (p *testIndexProvider) changeTrackStart() {
	p.changed = map[uint32]bool{}
}

// Stops recording which chunks are saved and returns the list of saved chunks
func (p *testIndexProvider) changeTrackStop() []int {
	changed := make([]int, 0)
	for c := range p.changed {
		changed = append(changed, int(c))
	}
	p.changed = nil
	return changed
}

// Returns the contents of a given sequence
func (p *testIndexProvider) getSequenceContents(sequence uint32) (res *pb.IndexFile, err error) {
	// Get the data and un-serialize it
	data, _, _, err := p.Get(context.Background(), sequence)
	if err != nil {
		return nil, err
	}
	res = &pb.IndexFile{}
	err = proto.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

/* Extra methods added to the object */

// DumpState prints the state of the object
// Used for development/debugging
func (i *Index) DumpState() {
	fmt.Println("############Index state:\n############")

	if i.elements != nil {
		fmt.Println("Elements:")
		for _, v := range i.elements {
			fmt.Println(v.Path)
		}
	} else {
		fmt.Println("Elements is nil")
	}

	fmt.Println("#####")

	if i.cacheTree != nil {
		fmt.Println("Tree:")
		i.cacheTree.Dump()
	} else {
		fmt.Println("Tree is nil")
	}

	fmt.Println("#####")

	if i.cacheFiles != nil {
		fmt.Println("Cache files:")
		for k, v := range i.cacheFiles {
			fmt.Println(k, " - ", v.Path)
		}
	} else {
		fmt.Println("Cache files is nil")
	}

	fmt.Println("#####")

	if len(i.deleted) > 0 {
		fmt.Println("Deleted:")
		for k, v := range i.deleted {
			fmt.Printf("#%d: %d\n", k, v)
		}
	} else {
		fmt.Println("Deleted is nil")
	}

	fmt.Print("############\n\n")
}

// Dump information about this node and all its children
// Used for debugging
func (n *IndexTreeNode) Dump() {
	n.dump(0)
}

func (n *IndexTreeNode) dump(indent int) {
	prefix := strings.Repeat(" ", indent*3)

	fmt.Println(prefix+"- Name:", n.Name)
	if n.File != nil {
		if n.File.Deleted {
			fmt.Println(prefix + "  Deleted file")
		} else {
			fileId, err := uuid.FromBytes(n.File.FileId)
			if err != nil {
				panic(err)
			}
			fmt.Println(prefix+"  File:", n.File.Path, "("+fileId.String()+")")
		}
	}
	if len(n.Children) == 0 {
		fmt.Println(prefix + "  Leaf node")
	} else {
		fmt.Println(prefix + "  Children:")
		for _, c := range n.Children {
			c.dump(indent + 1)
		}
	}
}

/* Utils */

// StringInSlice checks if a string is contained inside a slice of strings
func StringInSlice(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
