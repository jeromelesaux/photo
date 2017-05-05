package pdf

import (
	"bytes"
	"encoding/base64"
	"github.com/Sirupsen/logrus"
	"image"
	"image/png"
	"os"
	"testing"
)

func TestGeneratorPdf(t *testing.T) {
	images := make([]string, 0)
	images = append(images, imgToBase64("img1.jpg"))
	images = append(images, imgToBase64("img2.jpg"))
	images = append(images, imgToBase64("img3.jpg"))
	pdf := CreateFilePdfAlbum("unittest", images, Images3XPerPages)
	if pdf != "unittest.pdf" {
		t.Fatal("expected test.pdf and get " + pdf)
	}
}

func TestGeneratorPdf2(t *testing.T) {
	images := make([]string, 0)
	images = append(images, imgToBase64("img1.jpg"))
	images = append(images, imgToBase64("img2.jpg"))
	images = append(images, imgToBase64("img3.jpg"))
	images = append(images, imgToBase64("img1.jpg"))
	pdf := CreateFilePdfAlbum("unittest2", images, Images3XPerPages)
	if pdf != "unittest2.pdf" {
		t.Fatal("expected test.pdf and get " + pdf)
	}
}

func TestGeneratorPdf3(t *testing.T) {
	images := make([]string, 0)
	images = append(images, imgToBase64("img1.jpg"))
	images = append(images, imgToBase64("img2.jpg"))
	images = append(images, imgToBase64("img3.jpg"))
	images = append(images, imgToBase64("img1.jpg"))
	pdf := CreateFilePdfAlbum("unittest3", images, Images4XPerPages)
	if pdf != "unittest3.pdf" {
		t.Fatal("expected test.pdf and get " + pdf)
	}
}

func imgToBase64(path string) string {
	f, err := os.Open(path)
	if err != nil {
		logrus.Error("Error while retreiving image content with error " + err.Error())
		return ""
	}
	img, _, _ := image.Decode(f)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		logrus.Error("Error while retreiving image content with error " + err.Error())
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
