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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

// AzureStorage stores files on Azure Blob Storage
type AzureStorage struct {
	masterKey          []byte
	storageAccountName string
	storageContainer   string
	storagePipeline    pipeline.Pipeline
	storageURL         string
	dataPath           string
}

func (f *AzureStorage) Init(connection string) error {
	// Ensure the connection string is valid and extract the parts
	// connection mus start with "azureblob:" or "azure:"
	// Then it must contain the storage account container
	r := regexp.MustCompile("^(azureblob|azure):([a-z0-9][a-z0-9-]{2,62})$")
	match := r.FindStringSubmatch(connection)
	if match == nil || len(match) != 3 {
		return errors.New("invalid connection string for Azure Blob Storage")
	}
	f.storageContainer = match[2]

	// Get the storage account name and key from the environment
	name := os.Getenv("AZURE_STORAGE_ACCOUNT")
	key := os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	if name == "" || key == "" {
		return errors.New("environmental variables AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_ACCESS_KEY are not defined")
	}
	f.storageAccountName = name

	// Storage endpoint
	f.storageURL = fmt.Sprintf("https://%s.blob.core.windows.net/%s", f.storageAccountName, f.storageContainer)

	// Authenticate with Azure Storage
	credential, err := azblob.NewSharedKeyCredential(f.storageAccountName, key)
	if err != nil {
		return err
	}
	f.storagePipeline = azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			MaxTries: 3,
		},
	})

	return nil
}

func (f *AzureStorage) SetDataPath(path string) {
	f.dataPath = path
}

func (f *AzureStorage) SetMasterKey(key []byte) {
	f.masterKey = key
}

func (f *AzureStorage) GetInfoFile() (info *infofile.InfoFile, err error) {
	// Create the blob URL
	u, err := url.Parse(f.storageURL + "/_info.json")
	if err != nil {
		return
	}
	blockBlobURL := azblob.NewBlockBlobURL(*u, f.storagePipeline)

	// Download the file
	resp, err := blockBlobURL.Download(context.Background(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)
	if err != nil {
		if stgErr, ok := err.(azblob.StorageError); !ok {
			err = fmt.Errorf("network error while downloading the file: %s", err.Error())
		} else {
			// Blob not found
			if stgErr.Response().StatusCode == http.StatusNotFound {
				err = nil
				return
			}
			err = fmt.Errorf("azure Storage error while downloading the file: %s", stgErr.Response().Status)
		}
		return
	}
	body := resp.Body(azblob.RetryReaderOptions{
		MaxRetryRequests: 3,
	})
	defer body.Close()

	// Read the entire file
	data, err := ioutil.ReadAll(body)
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

func (f *AzureStorage) SetInfoFile(info *infofile.InfoFile) (err error) {
	// Encode the content as JSON
	data, err := json.Marshal(info)
	if err != nil {
		return
	}

	// Create the blob URL
	u, err := url.Parse(f.storageURL + "/_info.json")
	if err != nil {
		return
	}
	blockBlobURL := azblob.NewBlockBlobURL(*u, f.storagePipeline)

	// Upload
	_, err = azblob.UploadBufferToBlockBlob(context.Background(), data, blockBlobURL, azblob.UploadToBlockBlobOptions{})
	if err != nil {
		if stgErr, ok := err.(azblob.StorageError); !ok {
			return fmt.Errorf("network error while uploading the file: %s", err.Error())
		} else {
			return fmt.Errorf("Azure Storage error failed while uploading the file: %s", stgErr.Response().Status)
		}
	}

	return
}

func (f *AzureStorage) Get(name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
	return f.GetWithRange(context.Background(), name, out, metadataCb, "")
}

func (f *AzureStorage) GetWithContext(ctx context.Context, name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
	return f.GetWithRange(ctx, name, out, metadataCb, "")
}

func (f *AzureStorage) GetWithRange(ctx context.Context, name string, out io.Writer, metadataCb crypto.MetadataCb, rng *PackageRange) (found bool, tag interface{}, err error) {
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

	// Create the blob URL
	u, err := url.Parse(f.storageURL + "/" + folder + name)
	if err != nil {
		return
	}
	blockBlobURL := azblob.NewBlockBlobURL(*u, f.storagePipeline)

	/*
		// Check if we have a range
		var offset int64 = 0
		var count int64 = azblob.CountToEnd
		if rng != nil {
			offset = rng.StartBytes()
			count = rng.LengthBytes()
		}
	*/
	// Always download the first chunk + 256 byte that contain the header and the metadata
	// If we are requesting more chunks, download them too
	var offset int64 = 0
	var count int64 = azblob.CountToEnd
	if rng != nil && rng.Start == 0 {
		// Read one extra byte, so we know if the file continues (will cause us to discard some data)
		count = rng.LengthBytes() + 1
	}

	// Fetch the first chunk
	resp, err := blockBlobURL.Download(ctx, offset, count, azblob.BlobAccessConditions{}, false)
	if err != nil {
		if stgErr, ok := err.(azblob.StorageError); !ok {
			err = fmt.Errorf("network error while downloading the file: %s", err.Error())
		} else {
			// Blob not found
			if stgErr.Response().StatusCode == http.StatusNotFound {
				found = false
				err = nil
				return
			}
			err = fmt.Errorf("azure Storage error while downloading the file: %s", stgErr.Response().Status)
		}
		return
	}
	body := resp.Body(azblob.RetryReaderOptions{
		MaxRetryRequests: 3,
	})
	defer body.Close()

	// Check if the file exists but it's empty
	if resp.ContentLength() == 0 {
		found = false
		return
	}

	// Get the ETag
	tagObj := resp.ETag()
	tag = &tagObj

	// Get the header
	headerLen, wrappedKey, body, err := crypto.GetFileHeader(body)
	if err != nil {
		return
	}

	// Unwrap the file's key
	key, err := crypto.UnwrapKey(f.masterKey, wrappedKey)
	if err != nil {
		return
	}

	// Decrypt each package
	// If we have downloaded the first packages only, check how much data we need to discard
	maxRead := math.MaxInt64
	if rng != nil {
		maxRead = rng.LengthBytes()
	}

	return
}

func (f *AzureStorage) Set(name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error) {
	return f.SetWithContext(context.Background(), name, in, tag, metadata)
}

func (f *AzureStorage) SetWithContext(ctx context.Context, name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if name[0] != '_' {
		folder = f.dataPath + "/"
	}

	// Create the blob URL
	u, err := url.Parse(f.storageURL + "/" + folder + name)
	if err != nil {
		return nil, err
	}
	blockBlobURL := azblob.NewBlockBlobURL(*u, f.storagePipeline)

	// Encrypt the data and upload it
	pr, pw := io.Pipe()
	go func() {
		err := crypto.EncryptFile(pw, in, f.masterKey, metadata)
		if err != nil {
			panic(err)
		}
		pw.Close()
	}()

	// If we have a tag (ETag), we will allow the upload to succeed only if the tag matches
	// If there's no ETag, then the upload can succeed only if there's no file already

	// Access conditions for blob uploads: disallow the operation if the blob already exists
	// See: https://docs.microsoft.com/en-us/rest/api/storageservices/specifying-conditional-headers-for-blob-service-operations#Subheading1
	var accessConditions azblob.BlobAccessConditions
	if tag == nil {
		// Uploads can succeed only if there's no blob at that path yet
		accessConditions = azblob.BlobAccessConditions{
			ModifiedAccessConditions: azblob.ModifiedAccessConditions{
				IfNoneMatch: "*",
			},
		}
	} else {
		// Uploads can succeed only if the file hasn't been modified since we downloaded it
		accessConditions = azblob.BlobAccessConditions{
			ModifiedAccessConditions: azblob.ModifiedAccessConditions{
				IfMatch: *tag.(*azblob.ETag),
			},
		}
	}

	resp, err := azblob.UploadStreamToBlockBlob(ctx, pr, blockBlobURL, azblob.UploadStreamToBlockBlobOptions{
		BufferSize:       3 * 1024 * 1024,
		MaxBuffers:       2,
		AccessConditions: accessConditions,
	})
	if err != nil {
		if stgErr, ok := err.(azblob.StorageError); !ok {
			return nil, fmt.Errorf("network error while uploading the file: %s", err.Error())
		} else {
			return nil, fmt.Errorf("Azure Storage error failed while uploading the file: %s", stgErr.Response().Status)
		}
	}

	// Get the ETag
	tagObj := resp.ETag()
	tagOut = &tagObj

	return tagOut, nil
}

func (f *AzureStorage) Delete(name string, tag interface{}) (err error) {
	if name == "" {
		err = errors.New("name is empty")
		return
	}

	// If the file doesn't start with _, it lives in a sub-folder inside the data path
	folder := ""
	if name[0] != '_' {
		folder = f.dataPath + "/"
	}

	// Create the blob URL
	u, err := url.Parse(f.storageURL + "/" + folder + name)
	if err != nil {
		return
	}
	blockBlobURL := azblob.NewBlockBlobURL(*u, f.storagePipeline)

	// If we have a tag (ETag), we will allow the operation to succeed only if the tag matches
	// If there's no ETag, then it will always be allowed
	var accessConditions azblob.BlobAccessConditions
	if tag != nil {
		// Operation can succeed only if the file hasn't been modified since we downloaded it
		accessConditions = azblob.BlobAccessConditions{
			ModifiedAccessConditions: azblob.ModifiedAccessConditions{
				IfMatch: *tag.(*azblob.ETag),
			},
		}
	}

	// Delete the blob
	_, err = blockBlobURL.Delete(context.Background(), azblob.DeleteSnapshotsOptionInclude, accessConditions)
	return
}
