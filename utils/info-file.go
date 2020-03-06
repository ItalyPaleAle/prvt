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

package utils

import (
	"encoding/json"
	"errors"
)

// InfoFile is the content of the info file
type InfoFile struct {
	App     string `json:"app"`
	Version uint16 `json:"ver"`
}

// InfoCreate creates a new info file
func InfoCreate() ([]byte, error) {
	info := InfoFile{
		App:     "e2e",
		Version: 1,
	}
	data, err := json.Marshal(info)
	return data, err
}

// InfoVerify verifies the info file
func InfoVerify(data []byte) (*InfoFile, error) {
	if len(data) == 0 {
		return nil, errors.New("info file is empty")
	}

	// Parse the JSON data
	info := &InfoFile{}
	err := json.Unmarshal(data, info)
	if err != nil {
		return nil, err
	}

	// Check the contents
	if info.App != "e2e" {
		return nil, errors.New("invalid app name in info file")
	}
	if info.Version != 1 {
		return nil, errors.New("unsupported info file version")
	}

	return info, nil
}
