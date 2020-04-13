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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"

	"github.com/minio/minio-go"
)

// S3 stores files on a S3-compatible service
// This implementation does not rely on tags because S3 does not support conditional put requests
type S3 struct {
	masterKey  []byte
	client     *minio.Client
	bucketName string
	dataPath   string
}

func (f *S3) Init(connection string) error {
	// Ensure the connection string is valid and extract the parts
	// connection mus start with "s3:"
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
	var err error
	f.client, err = minio.New(endpoint, accessKeyId, secretAccessKey, tls)
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
	obj, err := f.client.GetObject(f.bucketName, "_info.json", minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

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
	_, err = f.client.PutObject(f.bucketName, "_info.json", buf, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return err
	}

	return
}

func (f *S3) Get(name string, out io.Writer, metadataCb crypto.MetadataCb) (found bool, tag interface{}, err error) {
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
	obj, err := f.client.GetObject(f.bucketName, folder+name, minio.GetObjectOptions{})
	if err != nil {
		return
	}

	// Check if the file exists but it's empty
	stat, err := obj.Stat()
	if err != nil || stat.Size == 0 {
		found = false
		return
	}

	// Decrypt the data
	err = crypto.DecryptFile(out, obj, f.masterKey, metadataCb)
	if err != nil {
		return
	}

	return
}

func (f *S3) Set(name string, in io.Reader, tag interface{}, metadata *crypto.Metadata) (tagOut interface{}, err error) {
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
	_, err = f.client.PutObject(f.bucketName, folder+name, pr, -1, minio.PutObjectOptions{})
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

	err = f.client.RemoveObject(f.bucketName, folder+name)

	return
}
