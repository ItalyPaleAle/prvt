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
	"e2e/crypto"
	"os"
)

// Inspired by https://github.com/spf13/afero/blob/v1.2.2/os.go
// © 2014 Steve Francia <spf@spf13.com>
// Licensed under Apache License, Version 2.0

// OsFs is a Fs implementation that uses functions provided by the os package
type OsFs struct{}

func (OsFs) Open(name string) (File, error) {
	// Open the input file
	in, err := os.Open(name)
	if in == nil {
		// while this looks strange, we need to return a bare nil (of type nil) not
		// a nil value of type *os.File or nil won't be nil
		return nil, err
	}

	// Output stream
	out := crypto.NewCryptoFile(name)
	err = crypto.DecryptFile(out, in, []byte("hello world"))
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (OsFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
