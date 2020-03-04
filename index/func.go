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
	"strings"

	"github.com/gofrs/uuid"
)

// Basename returns the path of a file
func Basename(path string) string {
	// The next line is never -1 (not found) since path must start with /
	index := strings.LastIndex(path, "/")
	folder := path[0:(index + 1)]

	return folder
}

// GenerateFileId generates a new ID for a file
func GenerateFileId() (string, error) {
	fileIdUuid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return fileIdUuid.String(), nil
}
