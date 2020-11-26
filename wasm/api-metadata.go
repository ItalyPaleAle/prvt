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

package main

import (
	"context"
	"fmt"
	"syscall/js"
)

// GetFileMetadata is a JS function that returns the metadata for a file
// This returns a JS Promise that resolves with a JS dictionary
// Arguments: masterKey (bytes), fileId (string)
func GetFileMetadata() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 2 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		masterKey := bytesFromJs(args[0])
		var fileId string = args[1].String()

		// Ensure the fileId isn't empty
		if len(fileId) < 1 {
			return jsError("FileID is empty")
		}

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// The code makes a network request, so running it in a goroutine
			go func() {
				// Request the metadata
				// This uses the cache if it's available
				_, _, _, _, metadata, err := getFileMetadata(context.Background(), masterKey, fileId)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Response object
				res := map[string]interface{}{
					"name":     metadata.Name,
					"mimeType": metadata.ContentType,
					"size":     metadata.Size,
				}
				resolve.Invoke(res)
			}()

			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}
