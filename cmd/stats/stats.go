package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tajtiattila/iphoto"
)

func main() {
	dir := flag.String("dir", "", "directory to stat")
	zip := flag.String("zip", "", "zip file to stat")
	flag.Parse()

	if *dir != "" {
		stats(iphoto.LoadIphotoDir(*dir))
	}
	if *zip != "" {
		stats(iphoto.LoadIphotoZip(*zip))
	}
}

func stats(db *iphoto.DB, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println("Photos:", len(db.Photo))
}
