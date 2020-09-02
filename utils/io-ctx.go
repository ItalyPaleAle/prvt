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
	"context"
	"io"
)

type ioctx func(p []byte) (int, error)

func (f ioctx) Write(p []byte) (n int, err error) {
	return f(p)
}
func (f ioctx) Read(p []byte) (n int, err error) {
	return f(p)
}

// WriterFuncWithContext returns a stream writer that supports a context
func WriterFuncWithContext(ctx context.Context, w io.Writer) ioctx {
	return func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return w.Write(p)
		}
	}
}

// ReaderFuncWithContext returns a stream reader that supports a context
func ReaderFuncWithContext(ctx context.Context, r io.Reader) ioctx {
	return func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return r.Read(p)
		}
	}
}
