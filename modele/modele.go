package modele

import (
	"sync"
	"time"
)

var (
	FILESIZE_BIG      = "big"
	FILESIZE_MEDIUM   = "medium"
	FILESIZE_LITTLE   = "little"
	ORIGIN_GOOGLE     = "www.google.com"
	ORIGIN_FLICKR     = "www.flickr.com"
	ORIGIN_LOCAL      = "localhost"
	ActionHistoryChan chan *ActionHistory
	ActionsHistory    []*ActionHistory
	ActionHistoryOnce sync.Once
)

func InitActionsHistory() {
	ActionHistoryOnce.Do(
		func() {
			ActionHistoryChan = make(chan *ActionHistory, 1)
			ActionsHistory = make([]*ActionHistory, 0)
			go func() {
				for a := range ActionHistoryChan {
					ActionsHistory = append(ActionsHistory, a)
				}
			}()

		})
}

func PostActionMessage(message string) {
	ActionHistoryChan <- &ActionHistory{Message: message,
		Date: time.Now().String()}
}

type ActionHistory struct {
	Date    string `json:"date"`
	Message string `json:"message"`
}

type RegisteredSlave struct {
	MachineId string `json:"machineId"`
	Ip        string `json:"ipv4"`
}

type PhotoInformations struct {
	Tags      map[string]string `json:"tags"`
	Md5Sum    string            `json:"md5sum"`
	Filename  string            `json:"filename"`
	Filepath  string            `json:"filepath"`
	Thumbnail string            `json:"thumbnail"`
}

func NewPhotoInformations() *PhotoInformations {
	return &PhotoInformations{
		Tags: make(map[string]string, 0),
	}
}

type PhotoResponse struct {
	Message   string               `json:"error_message,omitempty"`
	Origin    string               `json:"origin,omitempty"`
	Version   string               `json:"version"`
	MachineId string               `json:"machine"`
	Photos    []*PhotoInformations `json:"photos"`
}

var (
	Landscape = "landscape"
	Portrait  = "portrait"
)

type ExportRawPhoto struct {
	Filename      string `json:"filename"`
	Base64Content string `json:"base64_content"`
	Orientation   string `json:"orientation"`
}

func NewPhotoResponse(message string, version string, machineid string, photos []*PhotoInformations) *PhotoResponse {
	return &PhotoResponse{
		Message:   message,
		Version:   version,
		MachineId: machineid,
		Photos:    photos,
	}
}

type FolderToScan struct {
	MachineId string   `json:"machineid"`
	Folders   []string `json:"folders_toscan"`
}

type JSTreeAttribute struct {
	Opened   bool `json:"opened"`
	Disabled bool `json:"disabled"`
	Selected bool `json:"selected"`
}

func NewJSTreeAttribute() *JSTreeAttribute {
	return &JSTreeAttribute{Opened: false, Disabled: false, Selected: false}
}

type DirectoryItemResponse struct {
	Message          string                   `json:"error_message,omitempty"`
	Name             string                   `json:"text"`
	Path             string                   `json:"id"`
	Directories      []*DirectoryItemResponse `json:"children"`
	Parent           *DirectoryItemResponse   `json:"-"`
	Deep             int                      `json:"-"`
	MachineId        string                   `json:"machineid"`
	JstreeAttributes *JSTreeAttribute         `json:"state"`
}

//func (d *DirectoryItemResponse) Stop() bool {
//	if d != nil {
//		parent := d
//		for parent.Parent != nil {
//			parent = parent.Parent
//		}
//		if (d.Deep - parent.Deep) >= 4 {
//			return true
//		}
//
//	}
//	return false
//}

const VERSION = "1.0Beta"
