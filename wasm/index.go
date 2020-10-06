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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall/js"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/index"
	"github.com/ItalyPaleAle/prvt/utils"
)

// GetIndex is a JS function that returns a new RepoIndex object
// Arguments: masterKey (bytes)
func GetIndex() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		masterKey := bytesFromJs(args[0])

		// Ensure the master key isn't empty
		if masterKey == nil || len(masterKey) < 1 {
			return jsError("Master key is empty")
		}

		// Init the RepoIndex object
		ri := RepoIndex{}
		ri.Init(baseUrl, masterKey)

		// Return the JS dictionary with the functions in the index object
		return ri.JSValue()
	})
}

// RepoIndex is object, that will be exported as JavaScript object, containing methods to work with the index
type RepoIndex struct {
	index *index.Index
}

// Init the object and set the master key and base URL
func (i *RepoIndex) Init(baseURL string, masterKey []byte) {
	// Init the index if it's not already
	if i.index == nil {
		i.index = &index.Index{}
	}

	// Re-create the provider object for the index object so the cache is reset too
	provider := &IndexProviderHTTP{}
	provider.Init(baseURL, masterKey)
	i.index.SetProvider(provider)
}

// JSValue returns the js.Value with the dictionary with all the methods this object exposes to JavaScript
func (i *RepoIndex) JSValue() js.Value {
	return js.ValueOf(map[string]interface{}{
		"refresh":       i.Refresh(),
		"addFile":       i.AddFile(),
		"stat":          i.Stat(),
		"getFileByPath": i.GetFileByPath(),
		"getFileById":   i.GetFileById(),
		"deleteFile":    i.DeleteFile(),
		"listFolder":    i.ListFolder(),
	})
}

// Refresh exports a js.Func object for Index.Refresh
// Returns a JS Promise
// Arguments: force (bool)
func (i *RepoIndex) Refresh() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		force := args[0].Bool()

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// Run in a goroutine because this might make a blocking call
			go func() {
				err := i.index.Refresh(force)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Resolve the Promise
				resolve.Invoke()
				return
			}()
			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

// AddFile exports a js.Func object for Index.AddFile
// Returns a JS Promise
// Not yet implemented
func (i *RepoIndex) AddFile() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return jsError("This method has not been implemented yet")
	})
}

// Stat exports a js.Func object for Index.Stat
// Returns a JS Promise
// Arguments: none
func (i *RepoIndex) Stat() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 0 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// Run in a goroutine because this might make a blocking call
			go func() {
				stats, err := i.index.Stat()
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Resolve the Promise
				// Returns the object converted to map[string]interface{}, so it'll be converted to a JS dictionary
				resolve.Invoke(utils.Mapify(stats))
				return
			}()
			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

// GetFileByPath exports a js.Func object for Index.GetFileByPath
// Arguments: path (string)
func (i *RepoIndex) GetFileByPath() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		path := args[0].String()

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// Run in a goroutine because this might make a blocking call
			go func() {
				file, err := i.index.GetFileByPath(path)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Resolve the Promise
				// Returns the object converted to map[string]interface{}, so it'll be converted to a JS dictionary
				resolve.Invoke(utils.Mapify(file))
				return
			}()
			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

// GetFileById exports a js.Func object for Index.GetFileById
// Returns a JS Promise
// Arguments: fileId (string)
func (i *RepoIndex) GetFileById() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		fileId := args[0].String()

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// Run in a goroutine because this might make a blocking call
			go func() {
				file, err := i.index.GetFileById(fileId)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Resolve the Promise
				// Returns the object converted to map[string]interface{}, so it'll be converted to a JS dictionary
				resolve.Invoke(utils.Mapify(file))
				return
			}()
			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

// DeleteFile exports a js.Func object for Index.DeleteFile
// Returns a JS Promise
// Not yet implemented
// Arguments: path (string)
func (i *RepoIndex) DeleteFile() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return jsError("This method has not been implemented yet")

		/*if len(args) != 1 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		path := args[0].String()

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// Run in a goroutine because this might make a blocking call
			go func() {
				objects, paths, err := i.index.DeleteFile(path)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Resolve the Promise
				// Returns a dictionary with objects and paths
				resolve.Invoke(map[string]interface{}{
					"objects": objects,
					"paths":   paths,
				})
				return
			}()
			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)*/
	})
}

// ListFolder exports a js.Func object for Index.ListFolder
// Returns a JS Promise
// Arguments: path (string)
func (i *RepoIndex) ListFolder() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		path := args[0].String()

		// Handler for the Promise
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve := args[0]
			reject := args[1]

			// Run in a goroutine because this might make a blocking call
			go func() {
				list, err := i.index.ListFolder(path)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Returns the slice converted to []map[string]interface{}, so it'll be converted to a an array of JS dictionaries
				res := make([]interface{}, len(list))
				for i, e := range list {
					res[i] = utils.Mapify(e)
				}

				// Resolve the Promise
				resolve.Invoke(res)
				return
			}()
			return nil
		})

		// Create and return the Promise object
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

// IndexProviderHTTP provides access to the index file using HTTP requests as back-end
// This is used by the Wasm interface
type IndexProviderHTTP struct {
	Client    *http.Client
	baseURL   string
	masterKey []byte
}

// Init the object and set the master key and base URL
func (i *IndexProviderHTTP) Init(baseURL string, masterKey []byte) {
	i.baseURL = baseURL
	i.masterKey = masterKey
	if i.Client == nil {
		i.Client = http.DefaultClient
	}
}

// Get the index file
func (i *IndexProviderHTTP) Get(ctx context.Context) (data []byte, isJSON bool, tag interface{}, err error) {
	isJSON = false

	// Abort if no client
	if i.Client == nil {
		err = errors.New("HTTP client is not initialized")
		return
	}

	// Request the index file by making a HTTP call
	var httpReq *http.Request
	httpReq, err = http.NewRequestWithContext(ctx, "GET", i.baseURL+"/rawfile/_index", nil)
	if err != nil {
		return
	}
	var httpResp *http.Response
	httpResp, err = client.Do(httpReq)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	// Ensure we got a 200 status code
	if httpResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("invalid response status code: %d", httpResp.StatusCode)
		return
	}

	// Decrypt the file
	buf := &bytes.Buffer{}
	_, _, _, err = crypto.DecryptFile(ctx, buf, httpResp.Body, i.masterKey, func(metadata *crypto.Metadata, metadataSize int32) bool {
		// Check if we're decoding a legacy JSON file
		if metadata.ContentType == "application/json" {
			isJSON = true
		} else if metadata.ContentType != "application/protobuf" {
			err = errors.New("invalid Content-Type: " + metadata.ContentType)
		}
		return true
	})
	if err != nil {
		return
	}

	// Return the bytes from the buffer
	if buf.Len() == 0 {
		data = nil
	} else {
		data = buf.Bytes()
	}

	return
}

// Set the index file
func (i *IndexProviderHTTP) Set(ctx context.Context, data []byte, cacheTag interface{}) (newTag interface{}, err error) {
	// This is not yet implemented
	return nil, errors.New("IndexProviderHTTP.Set has not been implemented yet")
}
