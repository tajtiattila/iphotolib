package iphotolib

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/unicode/norm"
)

func readIphotoDB(lib *Lib, prefix string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	mustReadIphotoDB(lib, prefix)
	return nil
}

func mustReadIphotoDB(lib *Lib, prefix string) {
	db, err := sql.Open("sqlite3", ":memory:")
	panicOn(err)
	defer db.Close()

	n := filepath.ToSlash(filepath.Join(prefix, "Library.apdb"))
	_, err = db.Exec("ATTACH DATABASE '" + n + "' as L;")
	panicOn(err)

	n = filepath.ToSlash(filepath.Join(prefix, "Properties.apdb"))
	_, err = db.Exec(`ATTACH DATABASE "` + n + `" as P;`)
	panicOn(err)

	n = filepath.ToSlash(filepath.Join(prefix, "Faces.db"))
	_, err = db.Exec(`ATTACH DATABASE "` + n + `" as F;`)
	panicOn(err)

	// imageTimeZoneName always seems to be GMT

	// Query photos and their event and place relationships
	//  Path:     imagePath
	//  Date:     imageDate
	//  FileSize: fileSize
	//  FileName: fileName
	//  Name:     name
	//  Rating:   mainRating
	//  Hidden:   isHidden
	//  Flagged:  isFlagged
	//  Original: isOriginal
	//  InTrash:  isInTrash
	//  Event: modelId from L.RKFolder as E where projectUuid == E.uuid
	//  Place: placeId from L.RKPlaceForVersion as P where modelId == P.versionId
	rows, err := db.Query(`SELECT
			V.modelId,
			COALESCE((SELECT E.modelId FROM L.RKFolder AS E WHERE E.uuid=V.projectUuid), 0),
			COALESCE((SELECT Q.placeId FROM L.RKPlaceForVersion AS Q WHERE Q.versionId=V.modelId), 0),
			M.imagePath, V.imageDate, M.fileSize,
			V.fileName, COALESCE(V.name, ""),
			COALESCE((SELECT stringProperty from P.RKUniqueString as U
				WHERE U.modelId=(select stringId from P.RKIptcProperty as I
				WHERE I.versionId=V.modelId AND I.propertyKey='Caption/Abstract')), ""),
			V.mainRating, V.isHidden, V.isFlagged, V.isOriginal, V.isInTrash
		FROM L.RKVersion AS V
			INNER JOIN L.RKMaster AS M ON V.masterUuid=M.uuid
		ORDER BY V.imageDate;`)
	panicOn(err)
	defer rows.Close()
	lib.Photo = make(map[PhotoKey]Photo)
	lib.EventPhoto = make(map[EventKey][]PhotoKey)
	lib.PlacePhoto = make(map[PlaceKey][]PhotoKey)
	for rows.Next() {
		var photoId, eventId, placeId int
		var p Photo
		var timestamp time.Time
		var isHidden, isFlagged, isOriginal, isInTrash int
		panicOn(rows.Scan(
			&photoId, &eventId, &placeId,
			&p.Path, &timestamp, &p.FileSize,
			&p.FileName, &p.Name, &p.Desc,
			&p.Rating, &isHidden, &isFlagged, &isOriginal, &isInTrash))
		p.FileName = norm.NFC.String(p.FileName)
		p.Name = norm.NFC.String(p.Name)
		p.Path = norm.NFC.String(p.Path)
		p.Date = fixTimeStamp(timestamp)
		p.Hidden = isHidden != 0
		p.Flagged = isFlagged != 0
		p.Original = isOriginal != 0
		p.InTrash = isInTrash != 0
		pk, ek, lk := PhotoKey(photoId), EventKey(eventId), PlaceKey(placeId)
		p.Event = ek
		p.Place = lk
		p.dir = lib.dir
		lib.Photo[pk] = p
		lib.EventPhoto[ek] = append(lib.EventPhoto[ek], pk)
		lib.PlacePhoto[lk] = append(lib.PlacePhoto[lk], pk)
	}
	panicOn(rows.Err())

	// Query events
	//   Name:     name
	//   MinDate:  minImageDate
	//   MaxDate:  maxImageDate
	//   Hidden:   isHidden
	//   Favorite: isFavorite
	//   InTrash:  isInTrash
	rows, err = db.Query(`SELECT modelId,
			name, COALESCE(minImageDate+0, 0), COALESCE(maxImageDate+0, 0),
			isHidden, isFavorite, isInTrash
		FROM L.RKFolder;`)
	panicOn(err)
	defer rows.Close()
	lib.Event = make(map[EventKey]Event)
	for rows.Next() {
		var eventId int
		var name string
		var mind, maxd float64
		var isHidden, isFavorite, isInTrash int
		panicOn(rows.Scan(&eventId, &name, &mind, &maxd,
			&isHidden, &isFavorite, &isInTrash))
		lib.Event[EventKey(eventId)] = Event{
			Name:     norm.NFC.String(name),
			MinDate:  toTimeStamp(mind),
			MaxDate:  toTimeStamp(maxd),
			Hidden:   isHidden != 0,
			Favorite: isFavorite != 0,
			InTrash:  isInTrash != 0,
		}
	}
	panicOn(rows.Err())

	// Query places
	//   Name: defaultName
	//   Min:  minLatitude, minLongitude
	//   Max:  maxLatitude, maxLongitude
	//   Centroid: centroid
	rows, err = db.Query(`SELECT modelId,
			COALESCE(defaultName, ""),
			minLatitude, minLongitude,
			maxLatitude, maxLongitude,
			centroid
		FROM P.RKPlace;`)
	panicOn(err)
	defer rows.Close()
	lib.Place = make(map[PlaceKey]Place)
	for rows.Next() {
		var placeId int
		var centroid string
		var p Place
		panicOn(rows.Scan(&placeId, &p.Name,
			&p.Min.Lat, &p.Min.Lon,
			&p.Max.Lat, &p.Max.Lon,
			&centroid))
		p.Name = norm.NFC.String(p.Name)
		_, err = fmt.Sscanf(centroid, "%v,%v", &p.Centroid.Lat, &p.Centroid.Lon)
		if err != nil {
			p.Centroid.Lat = (p.Min.Lat + p.Max.Lat) / 2
			p.Centroid.Lon = (p.Min.Lon + p.Max.Lon) / 2
		}
		lib.Place[PlaceKey(placeId)] = p
	}
	panicOn(rows.Err())

	// Query faces
	//  Name:     name
	//  FullName: fullName
	//  Email:    email
	rows, err = db.Query(`SELECT faceKey,
			COALESCE(name, ""),
			COALESCE(fullName, ""),
			COALESCE(email, "")
		FROM F.RKFaceName;`)
	panicOn(err)
	defer rows.Close()
	lib.Face = make(map[FaceKey]Face)
	for rows.Next() {
		var faceId int
		var f Face
		panicOn(rows.Scan(&faceId, &f.Name, &f.FullName, &f.Email))
		f.Name = norm.NFC.String(f.Name)
		f.FullName = norm.NFC.String(f.FullName)
		f.Email = norm.NFC.String(f.Email)
		lib.Face[FaceKey(faceId)] = f
	}
	panicOn(rows.Err())

	// Query face-photo relations
	rows, err = db.Query(`SELECT versionId, faceKey FROM L.RKVersionFaceContent;`)
	panicOn(err)
	defer rows.Close()
	lib.FacePhoto = make(map[FaceKey][]PhotoKey)
	lib.PhotoFace = make(map[PhotoKey][]FaceKey)
	for rows.Next() {
		var photoId, faceId int
		panicOn(rows.Scan(&photoId, &faceId))
		pk, fk := PhotoKey(photoId), FaceKey(faceId)
		lib.FacePhoto[fk] = append(lib.FacePhoto[fk], pk)
		lib.PhotoFace[pk] = append(lib.PhotoFace[pk], fk)
	}
	panicOn(rows.Err())
}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}

// http://www.ipadforums.net/ipad-help/84074-ipad-internal-date-format.html
const unixOffset = 978307200 // 2001-01-01 00:00

func toTimeStamp(n float64) time.Time {
	s := int64(n)
	nano := int64((n - float64(s)) * 1e9)
	s += unixOffset
	return time.Unix(s, nano)
}

func fixTimeStamp(t time.Time) time.Time {
	// mattn/go-sqlite3 insists on returing TIMESTAMP as
	// time.Time values, but they are not UNIX timestamps
	// in iphoto.
	//
	// Fix such timestamps here.
	if t.IsZero() {
		return t
	}

	u := t.Unix() + unixOffset

	return time.Unix(u, 0)
}
