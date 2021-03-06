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

package repository

import (
	"errors"

	"github.com/ItalyPaleAle/prvt/fs"
	"github.com/ItalyPaleAle/prvt/index"
)

// Constants
const (
	// The method ignored the file
	RepositoryStatusIgnored = iota - 3
	// The file was already existing
	RepositoryStatusExisting
	// The file could not be found
	RepositoryStatusNotFound
	// Everything went well
	RepositoryStatusOK
	// An internal (application) error happened
	RepositoryStatusInternalError
	// An error occurred because of the user
	RepositoryStatusUserError
)

// PathResultMessage is the message passed to the res channel in AddPath/RemovePath
type PathResultMessage struct {
	Path   string
	Status int
	Err    error
	FileId string
}

// Repository is the object that manages the repository
type Repository struct {
	Store fs.Fs
	Index *index.Index

	tx index.IndexTxId
}

// BeginTransaction starts a transaction to add or remove multiple files at once
func (repo *Repository) BeginTransaction() error {
	if repo.Index == nil {
		return errors.New("index is not defined")
	}

	// Begin a transaction with the index
	repo.tx = repo.Index.BeginTransaction()
	return nil
}

// CommitTransaction commits a transaction and saves pending changes
func (repo *Repository) CommitTransaction() error {
	if repo.tx == 0 {
		return errors.New("no active transaction")
	}

	// Begin a transaction with the index
	err := repo.Index.CommitTransaction(repo.tx)
	repo.tx = 0
	return err
}
