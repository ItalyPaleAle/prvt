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

import "syscall/js"

// Wraps a message into a JavaScript object of type error
func jsError(message string) js.Value {
	errConstructor := js.Global().Get("Error")
	errVal := errConstructor.New(message)
	return errVal
}

// Returns a byte slice from a js.Value
func bytesFromJs(arg js.Value) []byte {
	out := make([]byte, arg.Length())
	js.CopyBytesToGo(out, arg)
	return out
}

// Returns a js.Value from a byte slice
func jsFromBytes(data []byte) js.Value {
	arrayConstructor := js.Global().Get("Uint8Array")
	result := arrayConstructor.New(len(data))
	js.CopyBytesToJS(result, data)
	return result
}
