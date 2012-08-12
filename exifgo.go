package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

const (
	MAX_HEADER_SIZE = 64 * 1024
	DELIM           = 0xff
	EOI             = 0xd9
	SOI_MARKER      = "\xff\xd8"
	EOI_MARKER      = "\xff\xd9"

	EXIF_OFFSET = 0x8769
	GPSIFD      = 0x8825

	TIFF_OFFSET = 6
	TIFF_TAG    = 0x2a
)

type marker struct {
	Label string
}

var jpeg_markers = map[byte]marker{
	0xc0: marker{"SOF0"},
	0xc2: marker{"SOF2"},
	0xc4: marker{"DHT"},
	0xda: marker{"SOS"},
	0xdb: marker{"DQT"},
	0xdd: marker{"DRI"},
	0xe0: marker{"APP0"},
	0xe1: marker{"APP1"},
	0xe2: marker{"APP2"},
	0xe3: marker{"APP3"},
	0xe4: marker{"APP4"},
	0xe5: marker{"APP5"},
	0xe6: marker{"APP6"},
	0xe7: marker{"APP7"},
	0xe8: marker{"APP8"},
	0xe9: marker{"APP9"},
	0xea: marker{"APP10"},
	0xeb: marker{"APP11"},
	0xec: marker{"APP12"},
	0xed: marker{"APP13"},
	0xee: marker{"APP14"},
	0xef: marker{"APP15"},
	0xfe: marker{"COM"},
}

var testimage = "/home/anders/IMG_0131.JPG"

func main() {
	file, err := os.Open(testimage)
	if err != nil {
		log.Fatal(err)
	}

	soi_marker := make([]byte, len(SOI_MARKER))
	_, err = file.Read(soi_marker)
	if err != nil {
		log.Fatal(err)
	}
	if string(soi_marker) != SOI_MARKER {
		log.Fatal("invalid image file. not a jpeg")
	}

	head := make([]byte, 2)
	var head2 uint16
	for {
		//             head = input.read(2)
		//             delim, mark  =  unpack(">BB", head)
		err = binary.Read(file, binary.BigEndian, head)
		delim := head[0]
		mark := head[1]
		//             if (delim != DELIM):
		//                 raise self.InvalidFile("Error, expecting delmiter. "\
		//                                        "Got <%s> should be <%s>" %
		//                                        (delim, DELIM))
		if delim != DELIM {
			break
		}
		//             if mark == EOI:
		//                 # Hit end of image marker, game-over!
		//                 break
		if mark == EOI {
			break
		}
		//             head2 = input.read(2)
		//             size = unpack(">H", head2)[0]
		err = binary.Read(file, binary.BigEndian, &head2)
		//             data = input.read(size-2)
		data := make([]byte, head2-2)
		_, err = file.Read(data)
		m, found := jpeg_markers[mark]
		if !found {
			continue
		}
		if m.Label == "APP1" {
			parse_exif(data)
		} 
	}
	// # Now go through and find all the blocks of data
	//         segments = []
	//         while 1:

	//         self._segments = segments

	//	data := make([]byte, 100)
	//	count, err = file.Read(data)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	fmt.Printf("read %d bytes: %q\n", count, data[:count])

}

func parse_exif(data []byte) {
	exif := data[:6]
	if string(exif) != "Exif\x00\x00" {
		return
	}
	tiff_data := data[TIFF_OFFSET:]
	tiff_endian := tiff_data[:2]
	var e binary.ByteOrder
	e = binary.LittleEndian
	if string(tiff_endian) == "II" {
	} else {
		if string(tiff_endian) == "MM" {
			e = binary.BigEndian
		} else {
			return
		}
	}
	//     tiff_tag, tiff_offset = unpack(self.e + 'HI', tiff_data[2:8])
	var tiff_tag uint16
	var tiff_offset uint32
	binary.Read(bytes.NewBuffer(tiff_data[2:4]), e, &tiff_tag)
	binary.Read(bytes.NewBuffer(tiff_data[4:8]), e, &tiff_offset)
	//     if (tiff_tag != TIFF_TAG):
	//         raise JpegFile.InvalidFile("Bad TIFF tag. Got <%x>, expecting "\
	//                                    "<%x>" % (tiff_tag, TIFF_TAG))

	if tiff_tag != TIFF_TAG {
	}
	//     # Ok, the header parse out OK. Now we parse the IFDs contained in
	//     # the APP1 header.

	//     # We use this loop, even though we can really only expect and support
	//     # two IFDs, the Attribute data and the Thumbnail data
	//     offset = tiff_offset
	//     count = 0
	offset := tiff_offset
	count := 0

	var num_entries uint16
	var start uint32
	//     while offset:
	for offset > 0 {
		//         count += 1
		count++
		//         num_entries = unpack(self.e + 'H', tiff_data[offset:offset+2])[0]
		binary.Read(bytes.NewBuffer(tiff_data[offset:offset+2]), e, &num_entries)
		//         start = 2 + offset + (num_entries*12)
		start = 2 + offset + (uint32(num_entries) * 12)

		//         if (count == 1):
		//             ifd = IfdTIFF(self.e, offset, self, self.mode, tiff_data)
		if count == 1 {
			ifdtiff(e, offset, tiff_data)
		} else if count == 2 {
			//         elif (count == 2):
			//             ifd = IfdThumbnail(self.e, offset, self, self.mode, tiff_data)
		} else {
			fmt.Println("invalid jpeg file")
		}
		//         self.ifds.append(ifd)

		//         # Get next offset
		//         offset = unpack(self.e + "I", tiff_data[start:start+4])[0]
		binary.Read(bytes.NewBuffer(tiff_data[start:start+4]), e, &offset)
	}

}

type tagdef struct {
	Label string
	Tag   string
}

var tags = map[uint32]tagdef{
	0x8769: tagdef{"Exif IFD Pointer", "ExifOffset"},
	0xA005: tagdef{"Interoparability IFD Pointer", "InteroparabilityIFD"},
	0x8825: tagdef{"GPS Info IFD Pointer", "GPSIFD"},
	0x100:  tagdef{"Image width", "ImageWidth"},
	0x101:  tagdef{"Image height", "ImageHeight"},
	0x102:  tagdef{"Number of bits per component", "BitsPerSample"},
	0x103:  tagdef{"Compression Scheme", "Compression"},
	0x106:  tagdef{"Pixel Composition", "PhotometricInterpretion"},
	0x112:  tagdef{"Orientation of image", "Orientation"},
	0x115:  tagdef{"Number of components", "SamplesPerPixel"},
	0x11c:  tagdef{"Image data arrangement", "PlanarConfiguration"},
	0x212:  tagdef{"Subsampling ration of Y to C", "YCbCrSubsampling"},
	0x213:  tagdef{"Y and C positioning", "YCbCrCoefficients"},
	0x11a:  tagdef{"X Resolution", "XResolution"},
	0x11b:  tagdef{"Y Resolution", "YResolution"},
	0x128:  tagdef{"Unit of X and Y resolution", "ResolutionUnit"},
	//         # B. Tags relating to recording offset
	0x111: tagdef{"Image data location", "StripOffsets"},
	0x116: tagdef{"Number of rows per strip", "RowsPerStrip"},
	0x117: tagdef{"Bytes per compressed strip", "StripByteCounts"},
	0x201: tagdef{"Offset to JPEG SOI", "JPEGInterchangeFormat"},
	0x202: tagdef{"Bytes of JPEG data", "JPEGInterchangeFormatLength"},
	//         # C. Tags relating to image data characteristics
	//         # D. Other tags
	0x132:  tagdef{"File change data and time", "DateTime"},
	0x10e:  tagdef{"Image title", "ImageDescription"},
	0x10f:  tagdef{"Camera Make", "Make"},
	0x110:  tagdef{"Camera Model", "Model"},
	0x131:  tagdef{"Camera Software", "Software"},
	0x13B:  tagdef{"Artist", "Artist"},
	0x8298: tagdef{"Copyright holder", "Copyright"},
	// Extended EXIF tags
	0x9000: tagdef{"Exif Version", "ExifVersion"},
	0xA000: tagdef{"Supported Flashpix version", "FlashpixVersion"},
	// B. Tag relating to Image Data Characteristics
	0xA001: tagdef{"Color Space Information", "ColorSpace"},
	// C. Tags relating to Image Configuration
	0x9101: tagdef{"Meaning of each component", "ComponentConfiguration"},
	0x9102: tagdef{"Image compression mode", "CompressedBitsPerPixel"},
	0xA002: tagdef{"Valid image width", "PixelXDimension"},
	0xA003: tagdef{"Valid image height", "PixelYDimension"},
	// D. Tags relatin to User informatio
	0x927c: tagdef{"Manufacturer notes", "MakerNote"},
	0x9286: tagdef{"User comments", "UserComment"},
	// E. Tag relating to related file information
	0xA004: tagdef{"Related audio file", "RelatedSoundFile"},
	// F. Tags relating to date and time
	0x9003: tagdef{"Date of original data generation", "DateTimeOriginal"},
	0x9004: tagdef{"Date of digital data generation", "DateTimeDigitized"},
	0x9290: tagdef{"DateTime subseconds", "SubSecTime"},
	0x9291: tagdef{"DateTime original subseconds", "SubSecTimeOriginal"},
	0x9292: tagdef{"DateTime digitized subseconds", "SubSecTimeDigitized"},
	// G. Tags relating to Picture taking conditions
	0x829a: tagdef{"Exposure Time", "ExposureTime"},
	0x829d: tagdef{"F Number", "FNumber"},
	0x8822: tagdef{"Exposure Program", "ExposureProgram"},
	0x8824: tagdef{"Spectral Sensitivity", "SpectralSensitivity"},
	0x8827: tagdef{"ISO Speed Rating", "ISOSpeedRatings"},
	0x8829: tagdef{"Optoelectric conversion factor", "OECF"},
	0x9201: tagdef{"Shutter speed", "ShutterSpeedValue"},
	0x9202: tagdef{"Aperture", "ApertureValue"},
	0x9203: tagdef{"Brightness", "BrightnessValue"},
	0x9204: tagdef{"Exposure bias", "ExposureBiasValue"},
	0x9205: tagdef{"Maximum lens apeture", "MaxApertureValue"},
	0x9206: tagdef{"Subject Distance", "SubjectDistance"},
	0x9207: tagdef{"Metering mode", "MeteringMode"},
	0x9208: tagdef{"Light mode", "LightSource"},
	0x9209: tagdef{"Flash", "Flash"},
	0x920a: tagdef{"Lens focal length", "FocalLength"},
	0x9214: tagdef{"Subject area", "Subject area"},
	0xa20b: tagdef{"Flash energy", "FlashEnergy"},
	0xa20c: tagdef{"Spatial frequency results", "SpatialFrquencyResponse"},
	0xa20e: tagdef{"Focal plane X resolution", "FocalPlaneXResolution"},
	0xa20f: tagdef{"Focal plane Y resolution", "FocalPlaneYResolution"},
	0xa210: tagdef{"Focal plane resolution unit", "FocalPlaneResolutionUnit"},
	0xa214: tagdef{"Subject location", "SubjectLocation"},
	0xa215: tagdef{"Exposure index", "ExposureIndex"},
	0xa217: tagdef{"Sensing method", "SensingMethod"},
	0xa300: tagdef{"File source", "FileSource"},
	0xa301: tagdef{"Scene type", "SceneType"},
	0xa302: tagdef{"CFA pattern", "CFAPattern"},
	0xa401: tagdef{"Customer image processing", "CustomerRendered"},
	0xa402: tagdef{"Exposure mode", "ExposureMode"},
	0xa403: tagdef{"White balance", "WhiteBalance"},
	0xa404: tagdef{"Digital zoom ratio", "DigitalZoomRation"},
	0xa405: tagdef{"Focal length in 35mm film", "FocalLengthIn35mmFilm"},
	0xa406: tagdef{"Scene capture type", "SceneCaptureType"},
	0xa407: tagdef{"Gain control", "GainControl"},
	0xa40a: tagdef{"Sharpness", "Sharpness"},
	0xa40c: tagdef{"Subject distance range", "SubjectDistanceRange"},
	// H. Other tags
	0xa420: tagdef{"Unique image ID", "ImageUniqueID"},
}

func ifdtiff(e binary.ByteOrder, offset uint32, tiff_data []byte) {
	// num_entries = unpack(e + 'H', data[offset:offset+2])[0]
	var num_entries uint16
	entries := make([]exifentry, 0)
	binary.Read(bytes.NewBuffer(tiff_data[offset:offset+2]), e, &num_entries)
	var embedded_tags = map[uint16]string{
		0xA005:      "interoperability",
		EXIF_OFFSET: "extendedEXIF",
		0x8825:      "GPS",
	}
	// for i in range(num_entries):
	for i := 0; i < int(num_entries); i++ {
		//     start = (i * 12) + 2 + offset
		start := (i * 12) + 2 + int(offset)
		//     entry = unpack(e + "HHII", data[start:start+12])
		//     tag, exif_type, components, the_data = entry
		var tag, exif_type uint16
		var components, the_data uint32
		var component_data []byte
		binary.Read(bytes.NewBuffer(tiff_data[start:start+2]), e, &tag)
		binary.Read(bytes.NewBuffer(tiff_data[start+2:start+4]), e, &exif_type)
		binary.Read(bytes.NewBuffer(tiff_data[start+4:start+8]), e, &components)
		binary.Read(bytes.NewBuffer(tiff_data[start+8:start+12]), e, &the_data)
		//     byte_size = exif_type_size(exif_type) * components
		byte_size := exif_type_size(exif_type) * components
		if et, ok := embedded_tags[tag]; ok {
			//         actual_data = self.embedded_tags[tag][1](e, the_data,
			//                                                  exif_file, self.mode, data)
			if et == "extendedEXIF" {
				ifdtiff(e, the_data, tiff_data)
			}
		} else {
			t, ok := tags[uint32(tag)]
			if !ok {
				t.Label = fmt.Sprintf("0x%x",uint32(tag))
			}
			if byte_size > 4 {
				//             the_data = data[the_data:the_data+byte_size]
				component_data = tiff_data[the_data : the_data+byte_size]
			} else {
				//             the_data = data[start+8:start+8+byte_size]
				component_data = tiff_data[start+8 : start+8+int(byte_size)]
			}
			//         if exif_type == BYTE or exif_type == UNDEFINED:
			//             actual_data = list(the_data)
			if exif_type == BYTE {
				fmt.Println("decoding BYTE data")
			} else if exif_type == ASCII {
				fmt.Printf("%s: %s\n", t.Label, string(component_data))
				//             if the_data[-1] != '\0':
				//                 actual_data = the_data + '\0'
				if component_data[len(component_data)-1] != 0x00 {
					fmt.Println("not null terminated")
					component_data = append(component_data, 0x00)
				}
			} else if exif_type == SHORT {
				//             actual_data = list(unpack(e + ("H" * components), the_data))
				var sdata uint16
				binary.Read(bytes.NewBuffer(component_data), e, &sdata)
				fmt.Printf("%s: %d\n", t.Label, sdata)
				
			} else if exif_type == LONG {
				var ldata uint32
				//             actual_data = list(unpack(e + ("I" * components), the_data))
				binary.Read(bytes.NewBuffer(component_data), e, &ldata)
				fmt.Printf("%s: %d\n", t.Label, ldata)
				
			} else if exif_type == SLONG {
				fmt.Println("decoding SLONG data")
				//             actual_data = list(unpack(e + ("i" * components), the_data))
			} else if exif_type == RATIONAL || exif_type == SRATIONAL {
				if exif_type == RATIONAL {
					var n,d uint32
					binary.Read(bytes.NewBuffer(component_data[0:4]),e,&n)
					binary.Read(bytes.NewBuffer(component_data[4:8]),e,&d)
					fmt.Printf("%s: %d / %d\n", t.Label, n, d)
				} else {
					var n,d int32
					binary.Read(bytes.NewBuffer(component_data[0:4]),e,&n)
					binary.Read(bytes.NewBuffer(component_data[4:8]),e,&d)
					fmt.Printf("%s: %d / %d\n", t.Label, n, d)
				}

				//             if exif_type == RATIONAL: t = "II"
				//             else: t = "ii"
				//             actual_data = []
				//             for i in range(components):
				//                 actual_data.append(Rational(*unpack(e + t,
				//                                                     the_data[i*8:
				//                                                              i*8+8])))
			} 
			//         self.special_handler(tag, actual_data)				
		}
		//     entry = (tag, exif_type, actual_data)
		//     self.entries.append(entry)
		entries = append(entries, exifentry{tag, exif_type, component_data})
	}
	// need this for handling thumbnails:
	// self.ifd_handler(data)
}

type exifentry struct {
	Tag        uint16
	Type       uint16
	ActualData []byte
}

type exiftype struct {
	Label string
	Size  uint32
}

const (
	BYTE      = 1
	ASCII     = 2
	SHORT     = 3
	LONG      = 4
	RATIONAL  = 5
	UNDEFINED = 7
	SLONG     = 9
	SRATIONAL = 10
)

var exif_type_lookup = map[uint16]exiftype{
	BYTE:      exiftype{"byte", 1},
	ASCII:     exiftype{"ascii", 1},
	SHORT:     exiftype{"short", 2},
	LONG:      exiftype{"long", 4},
	RATIONAL:  exiftype{"rational", 8},
	UNDEFINED: exiftype{"undefined", 1},
	SLONG:     exiftype{"slong", 4},
	SRATIONAL: exiftype{"srational", 8},
}

func exif_type_size(t uint16) uint32 {
	return exif_type_lookup[t].Size
}
