package iphotolib

import "time"

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

	dir photoDir
}
