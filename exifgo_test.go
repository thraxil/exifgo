package exifgo

import (
	"encoding/hex"
	_ "fmt"
	"log"
	"os"
	"testing"
)

var testimage = "test_images/pug.jpg"

type string_testcase struct {
	Label string
	Value string
}

var stringtestcases = []string_testcase{
	string_testcase{"Camera Make", "PENTAX Corporation "},
	string_testcase{"Camera Model", "PENTAX Optio S "},
	string_testcase{"Camera Software", "Optio S V1.00    "},
	string_testcase{"File change data and time", "2006:12:23 13:16:26"},
	string_testcase{"Date of original data generation", "2006:12:23 13:16:26"},
	string_testcase{"Date of digital data generation", "2006:12:23 13:16:26"},
}

/*
Orientation of image: 1
X Resolution: 72 / 1
Y Resolution: 72 / 1
Unit of X and Y resolution: 2
Y and C positioning: 1
Exposure Time: 1 / 8
F Number: 26 / 10
Exposure Program: 2
Image compression mode: 2048000 / 3145728
Exposure bias: 0 / 3
Maximum lens apeture: 28 / 10
Metering mode: 5
Light mode: 0
Flash: 16
Lens focal length: 580 / 100
Color Space Information: 1
Valid image width: 2048
Valid image height: 1536
Customer image processing: 0
Exposure mode: 0
White balance: 0
Digital zoom ratio: 0 / 0
Focal length in 35mm film: 35
Scene capture type: 0
Gain control: 0
0xa408: 0
0xa409: 0
Sharpness: 0
Subject distance range: 2
*/

func Test_Pug(t *testing.T) {
	file, err := os.Open(testimage)
	if err != nil {
		log.Fatal(err)
	}
	found_tags, err := parse_jpeg(file)
	if err != nil {
		log.Fatal(err)
	}
	if len(found_tags) == 0 {
		t.Error("no tags found")
	}
	for _, stc := range stringtestcases {
		found := false
		for _, f := range found_tags {
			if f.Label == stc.Label {
				found = true
				if f.Content.(string) != stc.Value+"\x00" {
					t.Error("Not the expected value")
					t.Error(f.Content.(string))
					t.Error(hex.EncodeToString([]byte(f.Content.(string))))
					t.Error(hex.EncodeToString([]byte(stc.Value + "\x00")))
				}
			}
		}
		if !found {
			t.Error("tag not found")
		}
	}
}
