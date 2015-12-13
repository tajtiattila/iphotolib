// Copyright (c) 2015 Attila Tajti
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package iphotolib

import (
	"io"
	"os"
	"time"
)

type Photo struct {
	Path     string
	Date     time.Time
	FileSize int64
	FileName string
	Name     string
	Desc     string
	Rating   int

	Event EventKey // Event this Photo belongs to
	Place PlaceKey // Place this Photo belongs to

	Hidden   bool
	Flagged  bool
	Original bool
	InTrash  bool

	// dir is user to open and stat image
	dir photoDir
}

// Stat returns the os.FileInfo structure describing the photo file.
func (p *Photo) Stat() (os.FileInfo, error) {
	return p.dir.Stat(p.Path)
}

// Open returns a io.ReadCloser to access the photo file contents.
func (p *Photo) Open() (io.ReadCloser, error) {
	return p.dir.Open("Masters/" + p.Path)
}

// OpenThumb returns a io.ReadCloser to access the contents of the tumbnail.
func (p *Photo) OpenThumb() (io.ReadCloser, error) {
	return p.dir.Open("Thumbnails/" + p.Path)
}
