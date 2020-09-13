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

type CtxReader func(p []byte) (n int, err error)

func (rf CtxReader) Read(p []byte) (n int, err error) { return rf(p) }

type CtxWriter func(p []byte) (n int, err error)

func (rf CtxWriter) Write(p []byte) (n int, err error) { return rf(p) }

// ReaderFuncWithContext returns a stream reader that supports a context
func ReaderFuncWithContext(ctx context.Context, r io.Reader) CtxReader {
	return func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return r.Read(p)
		}
	}
}

func CtxCopy(ctx context.Context, dst io.Writer, src io.Reader) (written int64, err error) {
	return io.Copy(dst, CtxReader(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return src.Read(p)
		}
	}))
}

func CtxCopyN(ctx context.Context, dst io.Writer, src io.Reader, n int64) (written int64, err error) {
	return io.CopyN(dst, CtxReader(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return src.Read(p)
		}
	}), n)
}
