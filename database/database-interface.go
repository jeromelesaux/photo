// package contains all the structures and functions to stores, queries the data from diedot database
package database

import (
	"github.com/jeromelesaux/photo/modele"
)

// interface of all mandatories functions
type DatabaseInterface interface {
	InsertNewData(response *modele.PhotoResponse) error
	QueryAll() ([]*DatabasePhotoRecord, error)
	QueryExtension(pattern string) ([]*DatabasePhotoRecord, error)
	QueryFilename(pattern string) ([]*DatabasePhotoRecord, error)
	QueryExifTag(pattern string, exiftag string) ([]*DatabasePhotoRecord, error)
	CleanDatabase() error
}

// structure of a photo record
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

// structure of an album record
type DatabaseAlbumRecord struct {
	AlbumName   string                 `json:"album_name"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Records     []*DatabasePhotoRecord `json:"records"`
}

// functions returns a new pointer of databaseAlbumRecord
func NewDatabaseAlbumRecord() *DatabaseAlbumRecord {
	return &DatabaseAlbumRecord{Records: make([]*DatabasePhotoRecord, 0)}
}

func NewDataseAlbumRecordWithData(data []*DatabasePhotoRecord) *DatabaseAlbumRecord {
	return &DatabaseAlbumRecord{Records: data}

}

// functions returns a new pointer of databasePhotoRecord
// with all data set from arguments of the function
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
