package main

import (
	"flag"
	"fmt"
	"os"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/tajtiattila/iphotolib"
)

var (
	dump bool
	size bool
)

func main() {
	flag.BoolVar(&dump, "dump", false, "dump more info")
	flag.BoolVar(&size, "size", false, "show image sizes in dump")
	flag.Parse()

	for _, a := range flag.Args() {
		stats(a)
	}
}

func stats(p string) {
	db, err := iphotolib.Open(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println("Photos:", len(db.Photo))
	if !dump {
		return
	}
	for _, p := range db.Photo {
		fmt.Println(p.Name, p.Path, p.FileName, sizeStr(&p))
	}
	fmt.Println("Events:", len(db.Event))
	for _, e := range db.Event {
		fmt.Println(e.Name, e.MinDate)
	}
	fmt.Println("Faces:", len(db.Face))
	for k, f := range db.Face {
		var x string
		if f.FullName != "" {
			x = "(" + f.FullName + ")"
		}
		fmt.Println(f.Name, x, len(db.FacePhoto[k]))
	}
}

func sizeStr(p *iphotolib.Photo) string {
	if !size {
		return ""
	}
	r, err := p.Open()
	if err != nil {
		return "ERROR " + err.Error()
	}
	defer r.Close()

	c, _, err := image.DecodeConfig(r)
	return fmt.Sprint(c.Width, "x", c.Height)
}
