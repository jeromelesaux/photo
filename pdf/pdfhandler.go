package pdf

import (
	"github.com/jung-kurt/gofpdf"
	"fmt"
	"math/rand"
	"github.com/Sirupsen/logrus"
	"image/png"
	"encoding/base64"
	"strings"
	"time"
	"os"
	"image/jpeg"
)

var (
	Images3XPerPages = "Images3XPerPages"
	Images4XPerPages = "Images4XPerPages"
)

func CreateAlbumPdf(albumName string, photos []string, imagesPerPage string) string {
	filesnames := saveImage(photos)
	
	f := gofpdf.New("P", "mm", "A4", "")
	switch imagesPerPage {
	case Images3XPerPages :
		for i:=0; i<len(filesnames);i+=3 {
			f = Add3Images(f,filesnames[i],filesnames[i+1],filesnames[i+2])
		}
	case Images4XPerPages:
		for i:=0; i<len(filesnames);i+=4 {
			f = Add4Images(f,filesnames[i],filesnames[i+1],filesnames[i+2],filesnames[i+3])
		}
	}

	if err := f.OutputFileAndClose(albumName+".pdf"); err!=nil {
		logrus.Infof("error while saving the album %s in pdf format with error %v", albumName, err.Error())
		removeImages(filesnames)
		return ""
	}

	removeImages(filesnames)

	return albumName+".pdf"
}

func removeImages(files []string) {
	for _,file := range files {
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
			logrus.Infof("error in creating temporary file %s with error %v", filename, err.Error())
		} else {
			defer f.Close()
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(content))
			img, err := png.Decode(reader)
			if err != nil {
				logrus.Infof("error in creating temporary file %s with error %v", filename, err.Error())
			} else {
				if err := jpeg.Encode(f,img,&jpeg.Options{Quality:99}); err != nil {
					logrus.Infof("error in encoding temporary file %s with error %v", filename, err.Error())
				}  else {
					files = append(files, filename)
					logrus.Infof("file %s saved",filename)
				}
			}

		}
	}
	return files
}

func Add4Images(f *gofpdf.Fpdf, img1, img2, img3, img4 string) *gofpdf.Fpdf {
	logrus.Info("save files into page "+img1+" " + img2+" "+img3+" "+img4)
	f.AddPage()
	f.SetDrawColor(255, 255, 255)
	f.Image(img1, 10, 60, 90, 0, false, "", 0, "")
	//pdf.SetDrawColor(0,35,102)
	f.Image(img2, 10, 150, 90, 0, false, "", 0, "")
	f.Image(img3, 110, 50, 90, 0, false, "", 0, "")
	f.Image(img4, 110, 150, 90, 0, false, "", 0, "")
	return f
}

func Add3Images(f *gofpdf.Fpdf, img1, img2, img3 string) *gofpdf.Fpdf {
	logrus.Info("save files into page "+img1+" " + img2+" "+img3)
	f.AddPage()
	f.SetDrawColor(255, 255, 255)
	f.Image(img1, 10, 60, 90, 0, true, "", 0, "")
	//pdf.SetDrawColor(0,35,102)
	f.Image(img2, 10, 150, 170, 0, true, "", 0, "")
	f.Image(img3, 110, 50, 90, 0, false, "", 0, "")
	return f
}