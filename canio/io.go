// Copyright 2017 can authors. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package canio

type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

type Reader interface {
	Read() (*Frame, error)
}

type Writer interface {
	Write(f *Frame) error
}

type Closer interface {
	Close() error
}
