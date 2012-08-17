package exifgo

import (
	"log"
	"os"
	"testing"
)

var testimage = "test_images/pug.jpg"

func Test_Pug(t *testing.T) {
	file, err := os.Open(testimage)
	if err != nil {
		log.Fatal(err)
	}
	err = parse_jpeg(file)
	if err != nil {
		log.Fatal(err)
	}
}
