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
	"errors"
	"os"
)

// PathExists returns true if the path exists on disk
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// IsRegularFile returns true if the path is a file
func IsRegularFile(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return false, err
	}
	switch mode := stat.Mode(); {
	case mode.IsDir():
		return false, nil
	case mode.IsRegular():
		return true, nil
	default:
		return false, errors.New("Invalid mode")
	}
}

// EnsureFolder creates a folder if it doesn't exist already
func EnsureFolder(path string) error {
	exists, err := PathExists(path)
	if err != nil {
		return err
	} else if !exists {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// TouchFile creates an empty file if it doesn't exist
func TouchFile(name string) (err error) {
	var file *os.File
	file, err = os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return
	}
	err = file.Close()
	return
}
