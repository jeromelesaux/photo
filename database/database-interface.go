package database

import (
	"photo/modele"
)

type DatabaseInterface interface {
	InsertNewData(response *modele.PhotoResponse) error
	QueryAll() ([]*DatabasePhotoRecord, error)
	QueryExtension(pattern string) ([]*DatabasePhotoRecord, error)
	QueryFilename(pattern string) ([]*DatabasePhotoRecord, error)
	QueryExifTag(pattern string, exiftag string) ([]*DatabasePhotoRecord, error)
	CleanDatabase() error
}

type DatabasePhotoRecord struct {
	Md5sum    string                 `json:"md5sum"`
	Type      string                 `json:"type"`
	Filename  string                 `json:"filename"`
	Filepath  string                 `json:"filepath"`
	ExifTags  map[string]interface{} `json:"exiftags"`
	Image     string                 `json:"image"`
	MachineId string                 `json:"machineid"`
	Thumbnail string                 `json:"thumbnail"`
}

type DatabaseAlbumRecord struct {
	AlbumName string                 `json:"album_name"`
	Records   []*DatabasePhotoRecord `json:"records"`
}

func NewDatabaseAlbumRecord() *DatabaseAlbumRecord {
	return &DatabaseAlbumRecord{Records: make([]*DatabasePhotoRecord, 0)}
}

func NewDatabasePhotoResponse(md5sum string, filename string, filepath string, machineid string, thumbnail string, exiftags map[string]interface{}) *DatabasePhotoRecord {
	return &DatabasePhotoRecord{
		Md5sum:    md5sum,
		Filename:  filename,
		Filepath:  filepath,
		MachineId: machineid,
		Image:     thumbnail,
		ExifTags:  exiftags,
	}
}
