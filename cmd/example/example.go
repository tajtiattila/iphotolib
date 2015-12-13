package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/tajtiattila/iphotolib"
)

func main() {
	for _, a := range os.Args[1:] {
		photoList(a)
	}
}

func photoList(path string) {
	lib, err := iphotolib.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	for _, p := range lib.Photo {
		c, err := imageConfig(&p)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(p.Path, c.Width, "x", c.Height)
		}
	}
}

func imageConfig(p *iphotolib.Photo) (image.Config, error) {
	r, err := p.Open()
	if err != nil {
		return image.Config{}, err
	}
	defer r.Close()

	c, _, err := image.DecodeConfig(r)
	return c, err
}
