package pdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jeromelesaux/photo/modele"
	"github.com/jung-kurt/gofpdf"
	logger "github.com/sirupsen/logrus"
)

var (
	Images3XPerPages = "Images3XPerPages"
	Images4XPerPages = "Images4XPerPages"
)

func CreatePdfAlbum(albumName string, photos []*modele.ExportRawPhoto, imagesPerPage string) []byte {
	filenames := make([]string, 0)
	f := gofpdf.New("P", "mm", "A4", "")
	defer f.Close()
	switch imagesPerPage {
	case Images3XPerPages:
		i := 0
		for i < len(photos) {
			var file1, file2, file3 string
			var or1, or2, or3 string
			if i < len(photos) {
				file1 = photos[i].Filename
				or1 = photos[i].Orientation
				filenames = append(filenames, photos[i].Filename)
			}
			i++
			if i < len(photos) {
				file2 = photos[i].Filename
				or2 = photos[i].Orientation
				filenames = append(filenames, photos[i].Filename)
			}
			i++
			if i < len(photos) {
				file3 = photos[i].Filename
				or3 = photos[i].Orientation
				filenames = append(filenames, photos[i].Filename)
			}
			i++
			f = Add3Images(f, file1, or1, file2, or2, file3, or3)
		}
	case Images4XPerPages:
		i := 0
		for i < len(photos) {
			var file1, file2, file3, file4 string
			var or1, or2, or3, or4 string
			if i < len(photos) {
				file1 = photos[i].Filename
				or1 = photos[i].Orientation
				filenames = append(filenames, photos[i].Filename)
			}
			i++
			if i < len(photos) {
				file2 = photos[i].Filename
				or2 = photos[i].Orientation
				filenames = append(filenames, photos[i].Filename)
			}
			i++
			if i < len(photos) {
				file3 = photos[i].Filename
				or3 = photos[i].Orientation
				filenames = append(filenames, photos[i].Filename)
			}
			i++
			if i < len(photos) {
				file4 = photos[i].Filename
				or4 = photos[i].Orientation
				filenames = append(filenames, photos[i].Filename)
			}
			i++
			f = Add4Images(f, file1, or1, file2, or2, file3, or3, file4, or4)
		}
	}

	b := new(bytes.Buffer)

	if err := f.Output(b); err != nil {
		logger.Infof("error while saving the album %s in pdf format with error %v", albumName, err.Error())
		removeImages(filenames)
		return b.Bytes()
	}

	removeImages(filenames)

	return b.Bytes()
}

func CreateFilePdfAlbum(albumName string, photos []*modele.ExportRawPhoto, imagesPerPage string) string {
	filesnames := saveImage(photos)

	f := gofpdf.New("P", "mm", "A4", "")
	defer f.Close()
	switch imagesPerPage {
	case Images3XPerPages:
		i := 0
		for i < len(filesnames) {
			var file1, file2, file3 string
			var or1, or2, or3 string
			if i < len(filesnames) {
				file1 = filesnames[i]
				or1 = photos[i].Orientation
			}
			i++
			if i < len(filesnames) {
				file2 = filesnames[i]
				or2 = photos[i].Orientation
			}
			i++
			if i < len(filesnames) {
				file3 = filesnames[i]
				or3 = photos[i].Orientation
			}
			i++
			f = Add3Images(f, file1, or1, file2, or2, file3, or3)
		}
	case Images4XPerPages:
		i := 0
		for i < len(filesnames) {
			var file1, file2, file3, file4 string
			var or1, or2, or3, or4 string
			if i < len(filesnames) {
				file1 = filesnames[i]
				or1 = photos[i].Orientation
			}
			i++
			if i < len(filesnames) {
				file2 = filesnames[i]
				or2 = photos[i].Orientation
			}
			i++
			if i < len(filesnames) {
				file3 = filesnames[i]
				or3 = photos[i].Orientation
			}
			i++
			if i < len(filesnames) {
				file4 = filesnames[i]
				or4 = photos[i].Orientation
			}
			i++
			f = Add4Images(f, file1, or1, file2, or2, file3, or3, file4, or4)
		}
	}

	if err := f.OutputFileAndClose(albumName + ".pdf"); err != nil {
		logger.Infof("error while saving the album %s in pdf format with error %v", albumName, err.Error())
		removeImages(filesnames)
		return ""
	}

	removeImages(filesnames)

	return albumName + ".pdf"
}

func removeImages(files []string) {
	for _, file := range files {
		os.Remove(file)
	}

}

func saveImage(photos []*modele.ExportRawPhoto) []string {
	files := make([]string, 0)
	rand.Seed(time.Now().UTC().UnixNano())
	for index, content := range photos {
		filename := fmt.Sprintf("img_%d_%d.jpg", index, rand.Int())
		f, err := os.Create(filename)
		if err != nil {
			logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
		} else {
			defer f.Close()
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(content.Base64Content))
			img, err := png.Decode(reader)
			if err != nil {
				logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
				os.Remove(filename)
			} else {
				if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 99}); err != nil {
					logger.Infof("error in encoding temporary file %s with error %v", filename, err.Error())
				} else {
					files = append(files, filename)
					logger.Infof("file %s saved", filename)
				}
			}

		}
	}
	return files
}

func Add4Images(f *gofpdf.Fpdf, img1, orientation1, img2, orientation2, img3, orientation3, img4, orientation4 string) *gofpdf.Fpdf {
	logger.Info("save files into page " + img1 + " " + img2 + " " + img3 + " " + img4)
	f.AddPage()
	f.Rect(0, 0, 210, 297, "F")
	f.SetFillColor(0, 0, 0)
	if img1 != "" {
		logger.Infof("%s:%s", img1, orientation1)
		if orientation1 == modele.Landscape {
			f.Image(img1, 10, 50, 90, 80, false, "", 0, "")
		} else {
			f.Image(img1, 20, 50, 65, 80, false, "", 0, "")
		}
	}
	if img2 != "" {
		logger.Infof("%s:%s", img2, orientation2)
		if orientation2 == modele.Landscape {
			f.Image(img2, 10, 180, 90, 80, false, "", 0, "")
		} else {
			f.Image(img2, 20, 180, 65, 80, false, "", 0, "")
		}
	}
	if img3 != "" {
		logger.Infof("%s:%s", img3, orientation3)
		if orientation3 == modele.Landscape {
			f.Image(img3, 110, 50, 90, 80, false, "", 0, "")
		} else {
			f.Image(img3, 120, 50, 65, 80, false, "", 0, "")
		}

	}
	if img4 != "" {
		logger.Infof("%s:%s", img4, orientation4)
		if orientation4 == modele.Landscape {
			f.Image(img4, 110, 180, 90, 80, false, "", 0, "")
		} else {
			f.Image(img4, 120, 180, 65, 80, false, "", 0, "")
		}
	}
	return f
}

func Add3Images(f *gofpdf.Fpdf, img1, orientation1, img2, orientation2, img3, orientation3 string) *gofpdf.Fpdf {
	logger.Info("save files into page " + img1 + " " + img2 + " " + img3)
	f.AddPage()
	f.Rect(0, 0, 210, 297, "F")
	f.SetFillColor(0, 0, 0)
	if img1 != "" {
		if orientation1 == modele.Landscape {
			f.Image(img1, 10, 50, 90, 65, false, "", 0, "")
		} else {
			f.Image(img1, 20, 50, 65, 75, false, "", 0, "")
		}
	}
	if img3 != "" {
		logger.Infof("%s:%s", img3, orientation3)
		if orientation3 == modele.Landscape {
			f.Image(img3, 110, 50, 90, 65, false, "", 0, "")
		} else {
			f.Image(img3, 120, 50, 65, 75, false, "", 0, "")
		}
	}
	if img2 != "" {
		logger.Infof("%s:%s", img2, orientation2)
		if orientation2 == modele.Landscape {

			f.Image(img2, 20, 150, 170, 120, false, "", 0, "")

		} else {
			f.Image(img2, 50, 150, 105, 130, false, "", 0, "")
		}
	}

	return f
}
