package pdf

import (
	"testing"
	"bytes"
	"image/png"
	"encoding/base64"
	"github.com/Sirupsen/logrus"
	"os"
	"image"
)

func TestGeneratorPdf(t *testing.T) {
	images := make([]string,0)
	images = append(images,imgToBase64("img1.jpg"))
	images = append(images,imgToBase64("img2.jpg"))
	images = append(images,imgToBase64("img3.jpg"))
	pdf := CreateAlbumPdf("unittest",images,Images3XPerPages)
	if pdf != "unittest.pdf" {
		t.Fatal("expected test.pdf and get "+pdf)
	}
}

func imgToBase64(path string )string {
	f,err :=os.Open(path)
	if err != nil {
		logrus.Error("Error while retreiving image content with error " + err.Error())
		return ""
	}
	img,_,_ := image.Decode(f)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		logrus.Error("Error while retreiving image content with error " + err.Error())
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}


//// This example demonstrates how images are included in documents.
//func ExampleFpdf_Image() {
//	pdf := gofpdf.New("P", "mm", "A4", "")
//	pdf.AddPage()
//	pdf.SetFont("Arial", "", 11)
//	pdf.Image(example.ImageFile("logo.png"), 10, 10, 30, 0, false, "", 0, "")
//	pdf.Text(50, 20, "logo.png")
//	pdf.Image(example.ImageFile("logo.gif"), 10, 40, 30, 0, false, "", 0, "")
//	pdf.Text(50, 50, "logo.gif")
//	pdf.Image(example.ImageFile("logo-gray.png"), 10, 70, 30, 0, false, "", 0, "")
//	pdf.Text(50, 80, "logo-gray.png")
//	pdf.Image(example.ImageFile("logo-rgb.png"), 10, 100, 30, 0, false, "", 0, "")
//	pdf.Text(50, 110, "logo-rgb.png")
//	pdf.Image(example.ImageFile("logo.jpg"), 10, 130, 30, 0, false, "", 0, "")
//	pdf.Text(50, 140, "logo.jpg")
//	fileStr := example.Filename("Fpdf_Image")
//	err := pdf.OutputFileAndClose(fileStr)
//	example.Summary(err, fileStr)
//	// Output:
//	// Successfully generated pdf/Fpdf_Image.pdf
//}
