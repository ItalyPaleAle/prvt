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

/*
Build with:
GOOS=js GOARCH=wasm go build -o  ../ui/dist/assets/app.wasm
brotli -9k ../ui/dist/assets/app.wasm

The Go WebAssembly runtime is at:
$GOROOT/misc/wasm/wasm_exec.js
*/

import (
	"fmt"
	"syscall/js"
)

// Package-wide variable containing the address of Prvt node
var baseUrl string

func main() {
	// Export a "Prvt" global object that contains our functions
	js.Global().Set("Prvt", map[string]interface{}{
		"setBaseURL":     SetBaseUrl(),
		"decryptRequest": DecryptRequest(),
		"getIndex":       GetIndex(),
		"unlock":         Unlock(),
	})

	// Prevent the function from returning, which is required in a wasm module
	select {}
}

// SetBaseUrl is a JS function that sets a new value for baseUrl
// Arguments: baseUrl (string)
func SetBaseUrl() js.Func {
	// JS Function
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return jsError(fmt.Sprintf("Invalid number of arguments passed: %d", len(args)))
		}
		baseUrl = args[0].String()
		return nil
	})
}
