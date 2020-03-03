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

import "os"

// AzureFs is a Fs implementation that uses Azure Blob Storage as backend
type AzureFs struct{}

func (AzureFs) Create(name string) (File, error) {
	f, e := os.Create(name)
	if f == nil {
		// while this looks strange, we need to return a bare nil (of type nil) not
		// a nil value of type *os.File or nil won't be nil
		return nil, e
	}
	return f, e
}

func (AzureFs) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (AzureFs) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (AzureFs) Open(name string) (File, error) {
	f, e := os.Open(name)
	if f == nil {
		// while this looks strange, we need to return a bare nil (of type nil) not
		// a nil value of type *os.File or nil won't be nil
		return nil, e
	}
	return f, e
}

func (AzureFs) Remove(name string) error {
	return os.Remove(name)
}

func (AzureFs) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (AzureFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
