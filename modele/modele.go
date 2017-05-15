package modele

var (
	FILESIZE_BIG    = "big"
	FILESIZE_MEDIUM = "medium"
	FILESIZE_LITTLE = "little"
	ORIGIN_GOOGLE   = "www.google.com"
	ORIGIN_FLICKR   = "www.flickr.com"
	LOCAL_ORIGINE   = "localhost"
)

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

type RawPhoto struct {
	Name string `json:"name,omitempty"`
	Data string `json:"data,omitempty"`
}

type PhotoResponse struct {
	Message   string               `json:"error_message,omitempty"`
	Origin    string               `json:"origin,omitempty"`
	Version   string               `json:"version"`
	MachineId string               `json:"machine"`
	Photos    []*PhotoInformations `json:"photos"`
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

const VERSION = "1.0Beta"
