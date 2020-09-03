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

package cmd

import (
	"fmt"
)

// Error types
const (
	ErrorApp  = "app"
	ErrorUser = "user"
)

// ExecError is a custom error object that implements the error interface
type ExecError struct {
	Type    string
	Message string
	Data    error
}

func (e ExecError) Error() string {
	prefix := ""
	switch e.Type {
	case ErrorApp:
		prefix = "[Fatal error] "
	case ErrorUser:
		prefix = "[Error] "
	}

	if e.Data != nil {
		return fmt.Sprintf("%s%s\n%s\n", prefix, e.Message, e.Data.Error())
	} else {
		return fmt.Sprintf("%s%s\n", prefix, e.Message)
	}
}

func (e ExecError) StatusCode() int {
	switch e.Type {
	case ErrorApp:
		return 2
	case ErrorUser:
		return 4
	default:
		return 1
	}
}

func NewExecError(errType string, errMessage string, errData error) error {
	return ExecError{
		Type:    errType,
		Message: errMessage,
		Data:    errData,
	}
}
