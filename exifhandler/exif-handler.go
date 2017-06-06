// package to get exif values from local images
package exifhandler

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/xiam/exif"
	"path"

	"bytes"
	"encoding/base64"
	"github.com/disintegration/imaging"
	"github.com/jeromelesaux/photo/hash"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jeromelesaux/photo/configurationexif"
	"github.com/jeromelesaux/photo/modele"
	"strings"
)

// function returns all exifs values ot a local image
// filepath is the file path of the image to treat.
func GetPhotoInformations(filePath string) (*modele.PhotoInformations, error) {
	abspath, err := filepath.Abs(filePath)
	filename := path.Base(filePath)
	thumbnail, _ := GetBase64Thumbnail(filePath)
	sum, err := hash.Md5Sum(filePath)
	if err != nil {
		return &modele.PhotoInformations{
			Filename:  filename,
			Filepath:  abspath,
			Md5Sum:    sum,
			Thumbnail: thumbnail,
		}, err
	}

	data, err := exif.Read(filePath)
	if err != nil {
		logger.Error(err.Error())
		return &modele.PhotoInformations{
			Filename:  filename,
			Filepath:  abspath,
			Md5Sum:    sum,
			Thumbnail: thumbnail,
		}, err
	}

	logger.Info("---------START----------")
	if data != nil {
		for key, val := range data.Tags {
			logger.Info(key + " = " + val)
		}
	}

	logger.Info("---------END----------")
	if err != nil {
		logger.Info(err.Error())
		return &modele.PhotoInformations{}, err
	}
	return &modele.PhotoInformations{
		Filename:  filename,
		Filepath:  abspath,
		Tags:      data.Tags,
		Md5Sum:    sum,
		Thumbnail: thumbnail,
	}, err
}

var Tags = make([]*modele.PhotoInformations, 0)

// function searches and returns all imformations of photos found in this local directory (directorypath)
// conf is the structure containing all images suffix to match.
func GetPhotosInformations(directorypath string, conf configurationexif.FileExtension) ([]*modele.PhotoInformations, error) {
	Tags = Tags[:0]
	err := filepath.Walk(directorypath, ScanExifFile(conf))
	return Tags, err

}

// file walker function to match all suffix from fileextention
func ScanExifFile(fileExtension configurationexif.FileExtension) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {

		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		if !info.IsDir() {
			f := filepath.Base(path)
			for _, d := range fileExtension.Extensions {
				if strings.HasSuffix(f, d) {
					logger.Info("Found file " + f)
					tags, _ := GetPhotoInformations(path)
					if tags.Filename != "" {
						Tags = append(Tags, tags)
					}
				}
			}
		}
		return nil
	}
}

// function returns the base64 content of the image path
func GetBase64Photo(path string) (string, string, error) {
	img, err := imaging.Open(path)
	if err != nil {
		logger.Error("Error while retreiving image content with error " + err.Error())
		return "", "", err
	}
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		logger.Error("Error while retreiving image content with error " + err.Error())
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), Orientation(img), nil
}

// function returns the base64 thumbnail of the image path
func GetBase64Thumbnail(path string) (string, error) {

	img, err := imaging.Open(path)
	if err != nil {
		logger.Error("Error while retreiving thumbnail with error " + err.Error())
		return "", err
	}
	var dst *image.NRGBA
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y
	if height > width {
		if height < 100 {
			thumb := imaging.Thumbnail(img, width, height, imaging.CatmullRom)
			dst = imaging.New(width, height, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
		} else {
			thumb := imaging.Thumbnail(img, 100, 100, imaging.CatmullRom)
			dst = imaging.New(100, 100, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
		}
	} else {
		if width < 100 {
			thumb := imaging.Thumbnail(img, width, height, imaging.CatmullRom)
			dst = imaging.New(width, height, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
		} else {
			thumb := imaging.Thumbnail(img, 100, 100, imaging.CatmullRom)
			dst = imaging.New(100, 100, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
		}
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, dst)
	if err != nil {
		logger.Error("Error while retreiving thumbnail with error " + err.Error())
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil

}

func Orientation(img image.Image) string {
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y
	if height > width {
		return modele.Portrait
	} else {
		return modele.Landscape
	}

}

// function returns the base64 content of the image url (web mode)
func GetBase64ThumbnailUrl(url string) (string, error) {
	client := &http.Client{}
	response, err := client.Get(url)
	if err != nil {
		logger.Errorf("Error while getting image from url %s with error %v", url, err)
		return "", err
	}
	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()

	img, err := imaging.Decode(response.Body)
	if err != nil {
		logger.Errorf("Error while retreiving thumbnail for url %s  with error %v", url, err)
		return "", err
	}
	var dst *image.NRGBA
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y
	if height > width {
		if height < 100 {
			thumb := imaging.Thumbnail(img, width, height, imaging.CatmullRom)
			dst = imaging.New(width, height, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
		} else {
			thumb := imaging.Thumbnail(img, 100, 100, imaging.CatmullRom)
			dst = imaging.New(100, 100, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
		}
	} else {
		if width < 100 {
			thumb := imaging.Thumbnail(img, width, height, imaging.CatmullRom)
			dst = imaging.New(width, height, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
		} else {
			thumb := imaging.Thumbnail(img, 100, 100, imaging.CatmullRom)
			dst = imaging.New(100, 100, color.NRGBA{0, 0, 0, 0})
			dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
		}
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, dst)
	if err != nil {
		logger.Error("Error while retreiving thumbnail with error " + err.Error())
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil

}
