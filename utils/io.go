package utils

import (
	"context"
	"io"
)

type CtxReader func(p []byte) (n int, err error)

func (rf CtxReader) Read(p []byte) (n int, err error) { return rf(p) }

type CtxWriter func(p []byte) (n int, err error)

func (rf CtxWriter) Write(p []byte) (n int, err error) { return rf(p) }

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
