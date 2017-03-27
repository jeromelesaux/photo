package exifhandler

import (
	"github.com/xiam/exif"
	"path"
	"photo/logger"

	"os"
	"path/filepath"
	"photo/hash"
	"photo/modele"
	"strings"
)

func GetPhotoInformations(filePath string) (*modele.TagsPhoto, error) {
	data, err := exif.Read(filePath)
	sum, err := hash.Md5Sum(filePath)
	abspath, err := filepath.Abs(filePath)
	filename := path.Base(filePath)
	if err != nil {
		logger.Log(err.Error())
		return &modele.TagsPhoto{
			Filename: filename,
			Filepath: abspath,
			Md5Sum:   sum,
		}, err
	}
	logger.Log("---------START----------")
	for key, val := range data.Tags {
		logger.Log(key + " = " + val)
	}

	logger.Log("---------END----------")
	if err != nil {
		logger.Log(err.Error())
		return &modele.TagsPhoto{}, err
	}
	return &modele.TagsPhoto{
		Filename: filename,
		Filepath: abspath,
		Tags:     data.Tags,
		Md5Sum:   sum,
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
			logger.Log(err.Error())
			return nil
		}
		if !info.IsDir() {
			f := filepath.Base(path)
			for _, d := range fileExtension.Extensions {
				if strings.HasSuffix(f, d) {
					logger.Log("Found file " + f)
					tags, _ := GetPhotoInformations(path)
					Tags = append(Tags, tags)
				}
			}
		}
		return nil
	}
}
