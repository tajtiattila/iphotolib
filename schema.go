package iphoto

import "time"

type Photo struct {
	Path     string
	Date     time.Time
	FileSize int64
	FileName string
	Name     string
	Desc     string
	Rating   int

	Event EventKey
	Place PlaceKey

	Hidden   bool
	Flagged  bool
	Original bool
	InTrash  bool
}

type PhotoKey int

type Event struct {
	Name    string
	MinDate time.Time
	MaxDate time.Time

	Hidden   bool
	Favorite bool
	InTrash  bool
}

type EventKey int

type Face struct {
	Name     string
	FullName string
	Email    string
}

type FaceKey int

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

type DB struct {
	Photo map[PhotoKey]Photo
	Event map[EventKey]Event
	Face  map[FaceKey]Face
	Place map[PlaceKey]Place

	EventPhoto map[EventKey][]PhotoKey
	FacePhoto  map[FaceKey][]PhotoKey
	PlacePhoto map[PlaceKey][]PhotoKey

	PhotoFace map[PhotoKey][]FaceKey
}
