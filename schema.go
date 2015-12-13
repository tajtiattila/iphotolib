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

import "time"

type PhotoKey int

// Event is an optionally named a time span within the library.
type Event struct {
	Name    string
	MinDate time.Time
	MaxDate time.Time

	Hidden   bool
	Favorite bool
	InTrash  bool
}

type EventKey int

// Face is someone or something featured on photos.
type Face struct {
	Name     string
	FullName string
	Email    string
}

type FaceKey int

// Place is the location where a photo was taken.
type Place struct {
	Name string

	Min LatLon
	Max LatLon

	Centroid LatLon
}

type PlaceKey int

type LatLon struct {
	Lat, Lon float64
}

// Lib contains all photos, events, faces and places
// found in the library, and the relationships between them.
type Lib struct {
	Photo map[PhotoKey]Photo
	Event map[EventKey]Event
	Face  map[FaceKey]Face
	Place map[PlaceKey]Place

	EventPhoto map[EventKey][]PhotoKey
	FacePhoto  map[FaceKey][]PhotoKey
	PlacePhoto map[PlaceKey][]PhotoKey

	PhotoFace map[PhotoKey][]FaceKey

	dir photoDir
}
