package database

import (
	"photo/modele"
	"strings"
)

type DatabaseMock struct {
	data []*DatabasePhotoResponse
}

var _ DatabaseInterface = (*DatabaseMock)(nil)

func NewDataBaseMock() (*DatabaseMock, error) {
	return &DatabaseMock{data: make([]*DatabasePhotoResponse, 0)}, nil
}

func (d *DatabaseMock) InsertNewData(response *modele.PhotoResponse) error {
	for _, item := range response.Photos {
		exifs := make(map[string]interface{}, 0)
		for tag, value := range item.Tags {
			exifs[tag] = value
		}

		toinsert := &DatabasePhotoResponse{
			Filename: item.Filename,
			Filepath: item.Filepath,
			Md5sum:   item.Md5Sum,
			ExifTags: exifs,
		}
		d.data = append(d.data, toinsert)
	}
	return nil
}
func (d *DatabaseMock) QueryAll() ([]*DatabasePhotoResponse, error) {
	return d.data, nil
}
func (d *DatabaseMock) QueryExtenstion(pattern string) ([]*DatabasePhotoResponse, error) {
	results := make([]*DatabasePhotoResponse, 0)
	for _, p := range d.data {
		if strings.Contains(p.Type, pattern) {
			results = append(results, p)
		}
	}
	return results, nil
}
func (d *DatabaseMock) QueryFilename(pattern string) ([]*DatabasePhotoResponse, error) {
	results := make([]*DatabasePhotoResponse, 0)
	for _, p := range d.data {
		if strings.Contains(p.Filename, pattern) {
			results = append(results, p)
		}
	}
	return results, nil
}
func (d *DatabaseMock) QueryExifTag(pattern string, exiftag string) ([]*DatabasePhotoResponse, error) {
	results := make([]*DatabasePhotoResponse, 0)
	for _, p := range d.data {

		if strings.Contains(p.ExifTags[exiftag].(string), pattern) {
			results = append(results, p)
		}
	}
	return results, nil

}
