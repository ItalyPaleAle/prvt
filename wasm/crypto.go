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
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall/js"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs/fsutils"
	"github.com/ItalyPaleAle/prvt/utils"
)

var metadataCache *fsutils.MetadataCache
var mux *sync.Mutex
var client *http.Client
var fileRequestIdRegexp = regexp.MustCompile("^\\/?(raw)?file\\/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})$")

// Init static variables
func init() {
	// Init the HTTP request client if needed
	if client == nil {
		client = &http.Client{}
	}

	// Init metadata cache if needed
	if metadataCache == nil {
		metadataCache = &fsutils.MetadataCache{}
		err := metadataCache.Init()
		if err != nil {
			panic(err)
		}
	}

	// Init the mutex if needed
	if mux == nil {
		mux = &sync.Mutex{}
	}
}

// DecryptRequest can be invoked by JS code to return a file and decrypt it in the client
// This returns a JS Promise that resolves with a JS Response object, which can be returned as response for a fetch request
// Supports both requests for full files and requests with a Range header
// Arguments: masterKey (bytes), req (JS Request object)
func DecryptRequest() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 2 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		masterKey := bytesFromJs(args[0])
		req := args[1]

		// Ensure req is a Request object
		requestConstructor := js.Global().Get("Request")
		if req.Type() != js.TypeObject || !req.InstanceOf(requestConstructor) {
			return jsError("Invalid type for req argument")
		}

		// URL for the request
		reqUrlVal := req.Get("url")
		if reqUrlVal.Type() != js.TypeString {
			return jsError("Empty or invalid URL from the request")
		}
		var reqUrlStr string = reqUrlVal.String()

		// Get the file ID
		reqUrl, err := url.Parse(reqUrlStr)
		if err != nil {
			return jsError("Cannot parse request URL: " + err.Error())
		}
		match := fileRequestIdRegexp.FindStringSubmatch(reqUrl.EscapedPath())
		if match == nil || len(match) != 3 || len(match[2]) == 0 {
			return jsError("File ID not found in the URL")
		}
		fileId := match[2]

		// Check if we have a range
		var rng *utils.HttpRange
		reqHeaders := req.Get("headers")
		if reqHeaders.Type() == js.TypeObject {
			rngHeaderVal := reqHeaders.Call("get", "Range")
			if rngHeaderVal.Type() == js.TypeString {
				rngHeader := rngHeaderVal.String()
				var err error
				rng, err = utils.ParseRange(rngHeader)
				if err != nil {
					return jsError("Cannot parse Range header: " + err.Error())
				}
			}
		}

		// Waiting for https://github.com/w3c/ServiceWorker/issues/1544
		/*// Check if we have an AbortSignal object we can use to listen to canceled request
		var ok bool = req.Get("signal").Truthy()
		if ok {
			fmt.Println("Has signal")
		}*/

		// Return a Promise
		// This is because HTTP request needs to be made in a separate goroutine: https://github.com/golang/go/issues/41310
		var method js.Func
		if rng == nil {
			// Full-file requests
			method = decryptRequestPromise(masterKey, fileId)
		} else {
			// Range requests
			method = decryptRangeRequestPromise(masterKey, fileId, rng)
		}
		promiseConstructor := js.Global().Get("Promise")
		promise := promiseConstructor.New(method)

		return promise
	})
}

// Returns the callback metadata; used by DecryptRequest
func requestMetadataCb(headers *js.Value, rng *fsutils.RequestRange, responseStatusCode *int, done chan int, cacheAdd *metadataCacheAdd) crypto.MetadataCbReturn {
	return func(metadata *crypto.Metadata, metadataSize int32) bool {
		// Contents-Type and Content-Disposition
		if metadata.ContentType != "" {
			headers.Call("set", "Content-Type", metadata.ContentType)
		} else {
			headers.Call("set", "Content-Type", "application/octet-stream")
		}
		contentDisposition := "inline"

		if metadata.Name != "" {
			fileName := strings.ReplaceAll(metadata.Name, "\"", "")
			contentDisposition += "; filename=\"" + fileName + "\""
		}
		headers.Call("set", "Content-Disposition", contentDisposition)

		// Handle range requests
		if rng != nil {
			if rng.Start >= rng.FileSize {
				*responseStatusCode = http.StatusRequestedRangeNotSatisfiable
			} else {
				// Spec: https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests
				// Content-Length is the length of the range itself
				headers.Call("set", "Content-Length", strconv.FormatInt(rng.Length, 10))
				headers.Call("set", "Content-Range", rng.ResponseHeaderValue())
				*responseStatusCode = http.StatusPartialContent
			}
		} else {
			// Content-Length and Accept-Ranges
			if metadata.Size > 0 {
				headers.Call("set", "Content-Length", strconv.FormatInt(metadata.Size, 10))
				headers.Call("set", "Accept-Ranges", "bytes")
			}
		}

		// Pass the metadata to the object that will cache it
		if cacheAdd != nil {
			cacheAdd.MetadataData(metadataSize, metadata)
		}

		// Signal that the metadata is ready
		if done != nil {
			done <- 1
		}

		return true
	}
}

// Returns a Promise that decrypts a full file
func decryptRequestPromise(masterKey []byte, fileId string) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Arguments: resolve (function), reject (function)
		resolve := args[0]
		reject := args[1]

		go func() {
			// Will be used to add metadata to the cache
			cacheAdd := &metadataCacheAdd{
				// File ID is the key for the cache
				Name: fileId,
			}

			// Context for the entire method
			ctx := context.Background()

			// Request the file by making a HTTP call
			httpReq, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"/rawfile/"+fileId, nil)
			if err != nil {
				reject.Invoke(jsError(err.Error()))
				return
			}
			httpResp, err := client.Do(httpReq)
			if err != nil {
				reject.Invoke(jsError(err.Error()))
				return
			}

			// Response status code
			responseStatusCode := httpResp.StatusCode

			// Create a Headers object, for the response
			headersConstructor := js.Global().Get("Headers")
			headers := headersConstructor.New()

			// Metadata callback for decrypting the file
			hasMetadata := make(chan int)
			metadataCb := requestMetadataCb(&headers, nil, &responseStatusCode, hasMetadata, cacheAdd)

			// Underlying source for the stream
			decryptFunc := func(out io.Writer, in io.Reader) (uint16, int32, []byte, error) {
				return crypto.DecryptFile(ctx, out, in, masterKey, metadataCb)
			}
			underlyingSource := responseUnderlyingSource(httpResp.Body, decryptFunc, cacheAdd)

			// Create a ReadableStream object
			readableStreamConstructor := js.Global().Get("ReadableStream")
			readableStream := readableStreamConstructor.New(underlyingSource)

			// Wait for the metadata to be available
			<-hasMetadata

			// Create the init argument for the Response constructor
			responseInitObj := map[string]interface{}{
				"status":     responseStatusCode,
				"statusText": http.StatusText(httpResp.StatusCode),
				"headers":    headers,
			}

			// Create a Response object with the stream inside
			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(readableStream, responseInitObj)

			resolve.Invoke(response)
		}()

		return nil
	})
}

// Returns a Promise that decrypts a range request
func decryptRangeRequestPromise(masterKey []byte, fileId string, rngHeader *utils.HttpRange) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Arguments: resolve (function), reject (function)
		resolve := args[0]
		reject := args[1]

		go func() {
			// Context for the entire method
			ctx := context.Background()

			// Get the range object
			rng := fsutils.NewRequestRange(rngHeader)

			// Create a Headers object, for the response
			headersConstructor := js.Global().Get("Headers")
			headers := headersConstructor.New()

			// Response status code (will be set later)
			responseStatusCode := http.StatusOK

			// Metadata callback
			metadataCb := requestMetadataCb(&headers, rng, &responseStatusCode, nil, nil)

			// Check if the metadata is in cache already
			mux.Lock()
			headerVersion, headerLength, wrappedKey, metadataLength, metadata := metadataCache.Get(fileId)
			mux.Unlock()
			if headerVersion == 0 || headerLength < 1 || wrappedKey == nil || len(wrappedKey) < 1 {
				// Need to request the metadata and cache it
				// For that, we need to request the header and the first package, which are at most 64kb + (32+256) bytes
				var length int64 = 64*1024 + 32 + 256
				mdHttpReq, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"/rawfile/"+fileId, nil)
				mdHttpReq.Header.Set("Range", fmt.Sprintf("bytes=0-%d", length))
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}
				mdHttpResp, err := client.Do(mdHttpReq)
				if err != nil {
					reject.Invoke(jsError(err.Error()))
					return
				}
				defer mdHttpResp.Body.Close()

				// Decrypt the data
				headerVersion, headerLength, wrappedKey, err = crypto.DecryptFile(ctx, nil, mdHttpResp.Body, masterKey, func(md *crypto.Metadata, sz int32) bool {
					metadata = md
					metadataLength = sz
					return false
				})
				if err != nil && err != crypto.ErrMetadataOnly {
					reject.Invoke(jsError(err.Error()))
					return
				}

				// Store the metadata in cache
				mux.Lock()
				metadataCache.Add(fileId, headerVersion, headerLength, wrappedKey, metadataLength, metadata)
				mux.Unlock()
			}

			// Add the offsets to the range object and set the file size (it's guaranteed it's set, or we wouldn't have a range request)
			rng.HeaderOffset = int64(headerLength)
			rng.MetadataOffset = int64(metadataLength)
			rng.SetFileSize(metadata.Size)

			// Invoke the metadata callback
			metadataCb(metadata, metadataLength)

			// Request the actual ranges that we need by making a HTTP call
			httpReq, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"/rawfile/"+fileId, nil)
			httpReq.Header.Set("Range", rng.RequestHeaderValue())
			if err != nil {
				reject.Invoke(jsError(err.Error()))
				return
			}
			httpResp, err := client.Do(httpReq)
			if err != nil {
				reject.Invoke(jsError(err.Error()))
				return
			}

			// Underlying source for the stream
			decryptFunc := func(out io.Writer, in io.Reader) (uint16, int32, []byte, error) {
				err := crypto.DecryptPackages(ctx, out, in, headerVersion, wrappedKey, masterKey, rng.StartPackage(), uint32(rng.SkipBeginning()), rng.Length, nil)
				return headerVersion, headerLength, wrappedKey, err
			}
			underlyingSource := responseUnderlyingSource(httpResp.Body, decryptFunc, nil)

			// Create a ReadableStream object
			readableStreamConstructor := js.Global().Get("ReadableStream")
			readableStream := readableStreamConstructor.New(underlyingSource)

			// Create the init argument for the Response constructor
			responseInitObj := map[string]interface{}{
				"status":     responseStatusCode,
				"statusText": http.StatusText(httpResp.StatusCode),
				"headers":    headers,
			}

			// Create a Response object with the stream inside
			responseConstructor := js.Global().Get("Response")
			response := responseConstructor.New(readableStream, responseInitObj)

			resolve.Invoke(response)
		}()

		return nil
	})
}

// Type for the decryptFunc parameter
type decryptFuncSignature func(out io.Writer, in io.Reader) (uint16, int32, []byte, error)

// Returns the "underlyingSource" object for the ReadableStream JS object sent with the response
func responseUnderlyingSource(body io.ReadCloser, decryptFunc decryptFuncSignature, cacheAdd *metadataCacheAdd) map[string]interface{} {
	// Underlying source for the stream
	// Specs: https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream/ReadableStream
	return map[string]interface{}{
		// start method
		"start": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// The first and only arg is the controller object
			controller := args[0]

			// Pipe for the response
			pr, pw := io.Pipe()

			// In background, send the response to the client
			go func() {
				for {
					buf := make([]byte, 65536)
					n, err := pr.Read(buf)
					fmt.Println("Read", n, err)
					if err != nil && err != io.EOF {
						// Stream had an error
						controller.Call("error", jsError(err.Error()))
						return
					}
					if n > 0 {
						controller.Call("enqueue", jsFromBytes(buf[0:n]))
					}
					if err == io.EOF {
						// Stream is done
						fmt.Println("Stream closed")
						controller.Call("close")
						return
					}
				}
			}()

			// Decrypt the stream in yet another background goroutine, as we can't block on a goroutine invoked by JS in wasm that is dealing with HTTP requests
			go func() {
				headerVersion, headerLength, wrappedKey, err := decryptFunc(pw, body)
				if err != nil {
					controller.Call("error", jsError(err.Error()))
					body.Close()
					pw.Close()
					return
				}
				body.Close()
				pw.Close()

				// Add to the cache
				if cacheAdd != nil {
					cacheAdd.HeaderData(headerVersion, headerLength, wrappedKey)
				}
			}()

			return nil
		}),
		// cancel method
		"cancel": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			fmt.Println("Method invoked: cancel")
			body.Close()

			return nil
		}),
	}
}

// Using metadataCacheAdd objects to store header and metadata in the cache, since they require two separate calls
// When both HeaderData and MetadataData are called, then the data is added to the cache
type metadataCacheAdd struct {
	Name string

	headerVersion  uint16
	headerLength   int32
	wrappedKey     []byte
	metadataLength int32
	metadata       *crypto.Metadata
	hasHeader      bool
	hasMetadata    bool
}

// HeaderData is invoked to store header information
func (m metadataCacheAdd) HeaderData(headerVersion uint16, headerLength int32, wrappedKey []byte) {
	m.headerVersion = headerVersion
	m.headerLength = headerLength
	m.wrappedKey = wrappedKey
	m.hasHeader = true
	m.step()
}

// MetadataData is invoked to store metadata
func (m metadataCacheAdd) MetadataData(metadataLength int32, metadata *crypto.Metadata) {
	m.metadataLength = metadataLength
	m.metadata = metadata
	m.hasMetadata = true
	m.step()
}

// Called by both HeaderData and MetadataData: when everything is set, adds the data to the cache
func (m metadataCacheAdd) step() {
	if m.hasHeader && m.hasMetadata {
		// Store the metadata in cache
		// Adding a lock here to prevent the case when adding this key causes the eviction of another one that's in use
		mux.Lock()
		metadataCache.Add(m.Name, m.headerVersion, m.headerLength, m.wrappedKey, m.metadataLength, m.metadata)
		mux.Unlock()
	}
}
