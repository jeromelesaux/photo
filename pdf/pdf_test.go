package pdf

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/jeromelesaux/photo/modele"
	"github.com/sirupsen/logrus"
)

func TestGeneratorPdf(t *testing.T) {
	images := make([]*modele.ExportRawPhoto, 0)
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img1.jpg"), Orientation: modele.Portrait})
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img2.jpg"), Orientation: modele.Portrait})
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img3.jpg"), Orientation: modele.Landscape})
	pdf := CreateFilePdfAlbum("unittest", images, Images3XPerPages)
	if pdf != "unittest.pdf" {
		t.Fatal("expected unittest.pdf and get " + pdf)
	}
}

func TestGeneratorPdf2(t *testing.T) {
	images := make([]*modele.ExportRawPhoto, 0)
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img1.jpg"), Orientation: modele.Portrait})
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img2.jpg"), Orientation: modele.Portrait})
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img3.jpg"), Orientation: modele.Landscape})
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img1.jpg"), Orientation: modele.Portrait})
	pdf := CreateFilePdfAlbum("unittest2", images, Images3XPerPages)
	if pdf != "unittest2.pdf" {
		t.Fatal("expected unittest2.pdf and get " + pdf)
	}
}

func TestGeneratorPdf3(t *testing.T) {
	images := make([]*modele.ExportRawPhoto, 0)
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img1.jpg"), Orientation: modele.Portrait})
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img2.jpg"), Orientation: modele.Portrait})
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img3.jpg"), Orientation: modele.Landscape})
	images = append(images, &modele.ExportRawPhoto{Base64Content: imgToBase64("img1.jpg"), Orientation: modele.Portrait})
	pdf := CreateFilePdfAlbum("unittest3", images, Images4XPerPages)
	if pdf != "unittest3.pdf" {
		t.Fatal("expected unittest3.pdf and get " + pdf)
	}
}

//func TestGeneratorPdfLandscapesPortraits3PerPages (t *testing.T) {
//	images := make([]*modele.ExportRawPhoto, 0)
//	images = append(images, &modele.ExportRawPhoto{Base64Content:imgToBase64("images/Portrait_1.jpg"), Orientation:modele.Portrait})
//	images = append(images, &modele.ExportRawPhoto{Base64Content:imgToBase64("images/Landscape_2.jpg"), Orientation:modele.Landscape})
//	images = append(images,&modele.ExportRawPhoto{Base64Content:imgToBase64("images/Portrait_2.jpg"), Orientation:modele.Portrait})
//	images = append(images, &modele.ExportRawPhoto{Base64Content:imgToBase64("images/Landscape_3.jpg"), Orientation:modele.Landscape})
//	images = append(images, &modele.ExportRawPhoto{Base64Content:imgToBase64("images/Portrait_3.jpg"), Orientation:modele.Portrait})
//	pdf := CreateFilePdfAlbum("orientation_test", images, Images3XPerPages)
//	if pdf != "orientation_test.pdf" {
//		t.Fatal("expected orientation_test.pdf and get " + pdf)
//	}
//}
//
//func TestGeneratorPdfLandscapesPortraits4PerPages (t *testing.T) {
//	images := make([]*modele.ExportRawPhoto, 0)
//	images = append(images, &modele.ExportRawPhoto{Base64Content:imgToBase64("images/Portrait_1.jpg"), Orientation:modele.Portrait})
//	images = append(images, &modele.ExportRawPhoto{Base64Content:imgToBase64("images/Landscape_2.jpg"), Orientation:modele.Landscape})
//	images = append(images,&modele.ExportRawPhoto{Base64Content:imgToBase64("images/Portrait_2.jpg"), Orientation:modele.Portrait})
//	images = append(images, &modele.ExportRawPhoto{Base64Content:imgToBase64("images/Landscape_3.jpg"), Orientation:modele.Landscape})
//	images = append(images, &modele.ExportRawPhoto{Base64Content:imgToBase64("images/Portrait_3.jpg"), Orientation:modele.Portrait})
//	pdf := CreateFilePdfAlbum("orientation_test4", images, Images4XPerPages)
//	if pdf != "orientation_test4.pdf" {
//		t.Fatal("expected orientation_test4.pdf and get " + pdf)
//	}
//}

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
