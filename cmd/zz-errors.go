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
	"io"
	"os"
)

// Error types
const (
	ErrorApp  = "app"
	ErrorUser = "user"
)

// Disables exiting the application in case of error; used by tests
var NoExitOnError bool = false

// ExitWithError prints and error then terminates the app
func ExitWithError(out io.Writer, errType string, errMessage string, errData error) {
	prefix := ""
	status := 1
	switch errType {
	case ErrorApp:
		prefix = "[Fatal error]"
		status = 2
	case ErrorUser:
		prefix = "[Error]"
		status = 4
	}

	if errData != nil {
		fmt.Fprintf(out, "%s %s\n%s\n", prefix, errMessage, errData.Error())
	} else {
		fmt.Fprintf(out, "%s %s\n", prefix, errMessage)
	}

	if !NoExitOnError {
		os.Exit(status)
	}
}
