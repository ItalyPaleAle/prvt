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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"syscall/js"

	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/keys"
)

// Unlock is a JS function that tries unlocking the repo with a passphrase and returns the master key
// Note that this doesn't support unlocking with a GPG key, because Wasm can't communicate with the user's GPG agent safely
// This returns a JS Promise that resolves with a JS dictionary
// Arguments: passphrase (string)
func Unlock() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		var passphrase string = args[0].String()

		// Ensure the passphrase isn't empty
		if len(passphrase) < 1 {
			return jsError("Passphrase is empty")
		}

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// The code makes a network request, so running it in a goroutine
			go func() {
				// Make a request for the info file
				httpResp, err := client.Get(baseUrl + "/api/repo/infofile")
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}
				defer httpResp.Body.Close()

				// Ensure we got a 200 status code
				if httpResp.StatusCode != http.StatusOK {
					err = fmt.Errorf("invalid response status code: %d", httpResp.StatusCode)
					return
				}

				// Parse the response
				respBody, err := ioutil.ReadAll(httpResp.Body)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}
				infofile := &infofile.InfoFile{}
				err = json.Unmarshal(respBody, infofile)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Try unlocking the repo
				masterKey, keyId, errMessage, err := keys.GetMasterKeyWithPassphrase(infofile, passphrase)
				if err != nil {
					reject.Invoke(jsError(err.Error() + ":" + errMessage))
					return
				}

				// Return the result
				resolve.Invoke(map[string]interface{}{
					"masterKey": jsFromBytes(masterKey),
					"keyId":     keyId,
				})
			}()

			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}
