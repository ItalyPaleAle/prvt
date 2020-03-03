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
	"os"
)

// Inspired by https://github.com/spf13/afero/blob/v1.2.2/os.go
// © 2014 Steve Francia <spf@spf13.com>
// Licensed under Apache License, Version 2.0

// OsFs is a Fs implementation that uses functions provided by the os package
type OsFs struct{}

func (OsFs) Create(name string) (File, error) {
	f, e := os.Create(name)
	if f == nil {
		// while this looks strange, we need to return a bare nil (of type nil) not
		// a nil value of type *os.File or nil won't be nil
		return nil, e
	}
	return f, e
}

func (OsFs) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (OsFs) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (OsFs) Open(name string) (File, error) {
	f, e := os.Open(name)
	if f == nil {
		// while this looks strange, we need to return a bare nil (of type nil) not
		// a nil value of type *os.File or nil won't be nil
		return nil, e
	}
	return f, e
}

func (OsFs) Remove(name string) error {
	return os.Remove(name)
}

func (OsFs) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (OsFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
