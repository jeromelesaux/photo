package exifhandler

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/xiam/exif"
	"path"

	"bytes"
	"encoding/base64"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"photo/hash"
	"photo/modele"
	"strings"
)

func GetPhotoInformations(filePath string) (*modele.TagsPhoto, error) {
	abspath, err := filepath.Abs(filePath)
	filename := path.Base(filePath)
	thumbnail, _ := GetBase64Thumbnail(filePath)
	sum, err := hash.Md5Sum(filePath)
	if err != nil {
		return &modele.TagsPhoto{
			Filename:  filename,
			Filepath:  abspath,
			Md5Sum:    sum,
			Thumbnail: thumbnail,
		}, err
	}

	data, err := exif.Read(filePath)
	if err != nil {
		logger.Error(err.Error())
		return &modele.TagsPhoto{
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
		return &modele.TagsPhoto{}, err
	}
	return &modele.TagsPhoto{
		Filename:  filename,
		Filepath:  abspath,
		Tags:      data.Tags,
		Md5Sum:    sum,
		Thumbnail: thumbnail,
	}, err
}

var Tags = make([]*modele.TagsPhoto, 0)

func GetPhotosInformations(directorypath string, conf modele.FileExtension) ([]*modele.TagsPhoto, error) {
	Tags = Tags[:0]
	err := filepath.Walk(directorypath, ScanExifFile(conf))
	return Tags, err

}

func ScanExifFile(fileExtension modele.FileExtension) filepath.WalkFunc {
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

func GetBase64Photo(path string) (string, error) {
	img, err := imaging.Open(path)
	if err != nil {
		logger.Error("Error while retreiving image content with error " + err.Error())
		return "", err
	}
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		logger.Error("Error while retreiving image content with error " + err.Error())
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

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
