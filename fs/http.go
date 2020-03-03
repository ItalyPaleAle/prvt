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
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Inspired by https://github.com/spf13/afero/blob/v1.2.2/httpFs.go
// © 2014 Steve Francia <spf@spf13.com>
// Licensed under Apache License, Version 2.0

type httpDir struct {
	basePath string
	fs       HttpFs
}

func (d httpDir) Open(name string) (http.File, error) {
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := string(d.basePath)
	if dir == "" {
		dir = "."
	}

	f, err := d.fs.Open(filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name))))
	if err != nil {
		return nil, err
	}
	return f, nil
}

type HttpFs struct {
	source Fs
}

func NewHttpFs(source Fs) *HttpFs {
	return &HttpFs{source: source}
}

func (h HttpFs) Dir(s string) *httpDir {
	return &httpDir{basePath: s, fs: h}
}

func (h HttpFs) Open(name string) (http.File, error) {
	f, err := h.source.Open(name)
	if err == nil {
		if httpfile, ok := f.(http.File); ok {
			return httpfile, nil
		}
	}
	return nil, err
}

func (h HttpFs) Stat(name string) (os.FileInfo, error) {
	return h.source.Stat(name)
}
