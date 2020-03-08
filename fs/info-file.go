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
)

// InfoFile is the content of the info file
type InfoFile struct {
	App              string `json:"app"`
	Version          uint16 `json:"ver"`
	Salt             []byte `json:"slt"`
	ConfirmationHash []byte `json:"ph"`
}

// InfoCreate creates a new info file
func InfoCreate(salt []byte, confirmationHash []byte) (*InfoFile, error) {
	info := &InfoFile{
		App:              "prvt",
		Version:          1,
		Salt:             salt,
		ConfirmationHash: confirmationHash,
	}
	return info, nil
}

// InfoValidate validates the info object
func InfoValidate(info *InfoFile) error {
	// Check the contents
	if info == nil {
		return errors.New("empty info object")
	}
	if info.App != "prvt" {
		return errors.New("invalid app name in info file")
	}
	if info.Version != 1 {
		return errors.New("unsupported info file version")
	}
	if len(info.Salt) != 16 {
		return errors.New("invalid salt in info file")
	}
	if len(info.ConfirmationHash) != 32 {
		return errors.New("invalid confirmation hash in info file")
	}

	return nil
}
