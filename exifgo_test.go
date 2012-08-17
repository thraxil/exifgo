package exifgo

import (
	"log"
	"os"
)

var testimage = "test_images/pug.jpg"

func main() {
	file, err := os.Open(testimage)
	if err != nil {
		log.Fatal(err)
	}
	err = parse_jpeg(file)
	if err != nil {
		log.Fatal(err)
	}
}
