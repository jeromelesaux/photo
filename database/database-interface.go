package database

import (
	"photo/modele"
)

type DatabaseInterface interface {
	InsertNewData(response *modele.PhotoResponse) error
	QueryAll() ([]*DatabasePhotoResponse, error)
	QueryExtension(pattern string) ([]*DatabasePhotoResponse, error)
	QueryFilename(pattern string) ([]*DatabasePhotoResponse, error)
	QueryExifTag(pattern string, exiftag string) ([]*DatabasePhotoResponse, error)
}

type DatabasePhotoResponse struct {
	Md5sum    string                 `json:"md5sum"`
	Type      string                 `json:"type"`
	Filename  string                 `json:"filename"`
	Filepath  string                 `json:"filepath"`
	ExifTags  map[string]interface{} `json:"exiftags"`
	Image     string                 `json:"image"`
	MachineId string                 `json:"machineid"`
}

func NewDatabasePhotoResponse(md5sum string, filename string, filepath string, machineid string, exiftags map[string]interface{}) *DatabasePhotoResponse {
	return &DatabasePhotoResponse{
		Md5sum:    md5sum,
		Filename:  filename,
		Filepath:  filepath,
		MachineId: machineid,
		ExifTags:  exiftags,
	}
}
