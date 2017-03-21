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

func GetPhotoInformations(filePath string) *modele.TagsPhoto {
	data, err := exif.Read(filePath)
	sum, _ := hash.Md5Sum(filePath)
	abspath, _ := filepath.Abs(filePath)
	filename := path.Base(filePath)
	if err != nil {
		logger.Log(err.Error())
		return &modele.TagsPhoto{
			Filename: filename,
			Filepath: abspath,
			Md5Sum:   sum,
		}
	}
	logger.Log("---------START----------")
	for key, val := range data.Tags {
		logger.Log(key + " = " + val)
	}

	logger.Log("---------END----------")
	if err != nil {
		logger.Log(err.Error())
		return &modele.TagsPhoto{}
	}
	return &modele.TagsPhoto{
		Filename: filename,
		Filepath: abspath,
		Tags:     data.Tags,
		Md5Sum:   sum,
	}
}

var Tags = make([]*modele.TagsPhoto, 0)

func GetPhotosInformations(directorypath string, conf modele.FileExtension) []*modele.TagsPhoto {
	Tags = Tags[:0]
	filepath.Walk(directorypath, ScanFile(conf))
	return Tags
}

func ScanFile(fileExtension modele.FileExtension) filepath.WalkFunc {
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
					Tags = append(Tags, GetPhotoInformations(path))
				}
			}
		}
		return nil
	}
}
