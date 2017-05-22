package pdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"github.com/jung-kurt/gofpdf"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	Images3XPerPages = "Images3XPerPages"
	Images4XPerPages = "Images4XPerPages"
)

func CreatePdfAlbum(albumName string, photos []string, imagesPerPage string) []byte {

	f := gofpdf.New("P", "mm", "A4", "")
	switch imagesPerPage {
	case Images3XPerPages:
		i := 0
		for i < len(photos) {
			var file1, file2, file3 string
			if i < len(photos) {
				file1 = photos[i]
			}
			i++
			if i < len(photos) {
				file2 = photos[i]
			}
			i++
			if i < len(photos) {
				file3 = photos[i]
			}
			i++
			f = Add3Images(f, file1, file2, file3)
		}
	case Images4XPerPages:
		i := 0
		for i < len(photos) {
			var file1, file2, file3, file4 string
			if i < len(photos) {
				file1 = photos[i]
			}
			i++
			if i < len(photos) {
				file2 = photos[i]
			}
			i++
			if i < len(photos) {
				file3 = photos[i]
			}
			i++
			if i < len(photos) {
				file4 = photos[i]
			}
			i++
			f = Add4Images(f, file1, file2, file3, file4)
		}
	}

	b := new(bytes.Buffer)

	if err := f.Output(b); err != nil {
		logger.Infof("error while saving the album %s in pdf format with error %v", albumName, err.Error())
		removeImages(photos)
		return b.Bytes()
	}

	removeImages(photos)

	return b.Bytes()
}

func CreateFilePdfAlbum(albumName string, photos []string, imagesPerPage string) string {
	filesnames := saveImage(photos)

	f := gofpdf.New("P", "mm", "A4", "")
	switch imagesPerPage {
	case Images3XPerPages:
		i := 0
		for i < len(filesnames) {
			var file1, file2, file3 string
			if i < len(filesnames) {
				file1 = filesnames[i]
			}
			i++
			if i < len(filesnames) {
				file2 = filesnames[i]
			}
			i++
			if i < len(filesnames) {
				file3 = filesnames[i]
			}
			i++
			f = Add3Images(f, file1, file2, file3)
		}
	case Images4XPerPages:
		i := 0
		for i < len(filesnames) {
			var file1, file2, file3, file4 string
			if i < len(filesnames) {
				file1 = filesnames[i]
			}
			i++
			if i < len(filesnames) {
				file2 = filesnames[i]
			}
			i++
			if i < len(filesnames) {
				file3 = filesnames[i]
			}
			i++
			if i < len(filesnames) {
				file4 = filesnames[i]
			}
			i++
			f = Add4Images(f, file1, file2, file3, file4)
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

func saveImage(photos []string) []string {
	files := make([]string, 0)
	rand.Seed(time.Now().UTC().UnixNano())
	for index, content := range photos {
		filename := fmt.Sprintf("img_%d_%d.jpg", index, rand.Int())
		f, err := os.Create(filename)
		if err != nil {
			logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
		} else {
			defer f.Close()
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(content))
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

func Add4Images(f *gofpdf.Fpdf, img1, img2, img3, img4 string) *gofpdf.Fpdf {
	logger.Info("save files into page " + img1 + " " + img2 + " " + img3 + " " + img4)
	f.AddPage()
	f.Rect(0, 0, 210, 297, "F")
	f.SetFillColor(0, 0, 0)
	if img1 != "" {
		f.Image(img1, 10, 50, 90, 80, false, "", 0, "")
	}
	if img2 != "" {
		f.Image(img2, 10, 150, 90, 80, false, "", 0, "")
	}
	if img3 != "" {
		f.Image(img3, 110, 50, 90, 80, false, "", 0, "")
	}
	if img4 != "" {
		f.Image(img4, 110, 150, 90, 80, false, "", 0, "")
	}
	return f
}

func Add3Images(f *gofpdf.Fpdf, img1, img2, img3 string) *gofpdf.Fpdf {
	logger.Info("save files into page " + img1 + " " + img2 + " " + img3)
	f.AddPage()
	f.Rect(0, 0, 210, 297, "F")
	f.SetFillColor(0, 0, 0)
	if img1 != "" {
		f.Image(img1, 10, 50, 90, 80, false, "", 0, "")
	}
	if img3 != "" {
		f.Image(img3, 110, 50, 90, 80, false, "", 0, "")
	}
	if img2 != "" {
		f.Image(img2, 20, 150, 170, 130, false, "", 0, "")
	}

	return f
}
