// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3 with static-linking exception.
// See LICENCE file for details.

package ratelimit

import (
	"io"
)

type reader struct {
	r      io.Reader
	bucket BucketI
}

// Reader returns a reader that is rate limited by
// the given token BucketI. Each token in the BucketI
// represents one byte.
func Reader(r io.Reader, bucket BucketI) io.Reader {
	return &reader{
		r:      r,
		bucket: bucket,
	}
}

func (r *reader) Read(buf []byte) (int, error) {
	n, err := r.r.Read(buf)
	if n <= 0 {
		return n, err
	}
	r.bucket.Wait(int64(n))
	return n, err
}

type writer struct {
	w      io.Writer
	bucket BucketI
}

// Writer returns a reader that is rate limited by
// the given token BucketI. Each token in the BucketI
// represents one byte.
func Writer(w io.Writer, bucket BucketI) io.Writer {
	return &writer{
		w:      w,
		bucket: bucket,
	}
}

func (w *writer) Write(buf []byte) (int, error) {
	w.bucket.Wait(int64(len(buf)))
	return w.w.Write(buf)
}
