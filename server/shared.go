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

package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TreeOperationResponse struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	FileId string `json:"fileId"`
	Error  string `json:"error,omitempty"`
}

type MetadataResponse struct {
	FileId   string     `json:"fileId"`
	Folder   string     `json:"folder"`
	Name     string     `json:"name"`
	Date     *time.Time `json:"date,omitempty"`
	MimeType string     `json:"mimeType,omitempty"`
	Size     int64      `json:"size,omitempty"`
	Digest   []byte     `json:"digest,omitempty"`
}

type RepoKeyListResponse struct {
	Keys []RepoKeyListItem `json:"keys"`
}

type RepoKeyListItem struct {
	KeyId string `json:"keyId"`
	Type  string `json:"type"`
	UID   string `json:"uid,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type UnlockKeyRequest struct {
	Type       string `json:"type" form:"type"`
	Passphrase string `json:"passphrase" form:"passphrase"`
}

type InfoResponse struct {
	RepoInfoResponse

	Name         string `json:"name"`
	AppVersion   string `json:"version,omitempty"`
	BuildID      string `json:"buildId,omitempty"`
	BuildTime    string `json:"buildTime,omitempty"`
	CommitHash   string `json:"commitHash,omitempty"`
	Runtime      string `json:"runtime,omitempty"`
	Info         string `json:"info,omitempty"`
	ReadOnly     bool   `json:"readOnly,omitempty"`
	RepoSelected bool   `json:"repoSelected"`
	RepoUnlocked bool   `json:"repoUnlocked"`
}

type RepoInfoResponse struct {
	StoreType    string `json:"storeType,omitempty"`
	StoreAccount string `json:"storeAccount,omitempty"`
	RepoID       string `json:"repoId,omitempty"`
	RepoVersion  uint16 `json:"repoVersion,omitempty"`
	FileCount    int    `json:"files,omitempty"`
	GPGUnlock    bool   `json:"gpgUnlock,omitempty"`
}

type ConnectionListItem struct {
	Type    string `json:"type"`
	Account string `json:"account"`
}

type ConnectionList map[string]ConnectionListItem

// FromBody adds data to the object from a request
func (p *UnlockKeyRequest) FromBody(c *gin.Context) (ok bool) {
	// Get the information to unlock the repository from the body
	if err := c.Bind(p); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Could not parse request body"})
		return false
	}

	// Validate the body
	if p.Type != "passphrase" && p.Type != "gpg" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Parameter 'type' must be either 'passphrase' or 'gpg'"})
		return false
	}
	if p.Type == "passphrase" && len(p.Passphrase) < 1 {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Parameter 'passphrase' must be set when 'type' is 'passphrase'"})
		return false
	}

	return true
}

type AddKeyRequest struct {
	Passphrase string `json:"passphrase" form:"passphrase"`
	GPGKeyId   string `json:"gpg" form:"gpg"`
}

// FromBody adds data to the object from a request
func (p *AddKeyRequest) FromBody(c *gin.Context) (ok bool) {
	// Get the content from the body
	if err := c.Bind(p); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"Could not parse request body"})
		return false
	}

	if (p.Passphrase == "" && p.GPGKeyId == "") || (p.Passphrase != "" && p.GPGKeyId != "") {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{"One and only one of `passphrase` and `gpg` must be set"})
		return false
	}

	return true
}
