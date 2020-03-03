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

package crypto

import (
	"bytes"
	"os"
)

// CryptoFile implements the fs.File interface
type CryptoFile struct {
	name string
	buf  *bytes.Buffer
}

func NewCryptoFile(name string) CryptoFile {
	buf := &bytes.Buffer{}
	obj := CryptoFile{
		name: name,
		buf:  buf,
	}
	return obj
}

func (f CryptoFile) Read(p []byte) (n int, err error) {
	return f.buf.Read(p)
}

func (f CryptoFile) Write(p []byte) (n int, err error) {
	return f.buf.Write(p)
}

func (f CryptoFile) Name() string {
	return f.name
}

func (f CryptoFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f CryptoFile) Readdirnames(n int) ([]string, error) {
	res := make([]string, 0)
	return res, nil
}

func (f CryptoFile) Stat() (os.FileInfo, error) {
	return nil, nil
}
