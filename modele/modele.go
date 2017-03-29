package modele

import (
	"encoding/json"
	"os"
	"photo/logger"
	"strings"
	"sync"
)

type Configuration struct {
	PhotoExifUrl    string `json:"photoexif_url"`
	PhotoExifPort   int    `json:"photoexif_port"`
	PhotoExifAction string `json:"photoexif_action"`
	DatabasePath string `json:"database_path"`
}

type TagsPhoto struct {
	Tags     map[string]string `json:"tags"`
	Md5Sum   string            `json:"md5sum"`
	Filename string            `json:"filename"`
	Filepath string            `json:"filepath"`
}

type PhotoResponse struct {
	Message string       `json:"error_message,omitempty"`
	Version string       `json:"version"`
	Photos  []*TagsPhoto `json:"photos"`
}

type FileExtension struct {
	Extensions []string
}

type FolderToScan struct {
	Folders []string `json:"folders_toscan"`
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
	JstreeAttributes *JSTreeAttribute         `json:"state"`
}

const VERSION = "1.0Beta"

var confFileExtensionMut sync.Mutex
var confPhotoExifMut sync.Mutex
var configuration *Configuration

func LoadPhotoExifConfiguration(configurationPathfile string) *Configuration {

	confPhotoExifMut.Lock()
	file, errOpen := os.Open(configurationPathfile)
	if errOpen != nil {
		logger.Log("Error while opening file " + configurationPathfile + " with error :" + errOpen.Error())
	}
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		logger.LogLn("error:", err)
	}
	logger.LogLn(*configuration)
	confPhotoExifMut.Unlock()
	return configuration
}

func GetConfiguration() *Configuration {
	return configuration
}
func LoadConfiguration(configurationFile string) FileExtension {
	configuration := FileExtension{}
	confFileExtensionMut.Lock()
	file, errOpen := os.Open(configurationFile)
	if errOpen != nil {
		logger.Log("Error while opening file " + configurationFile + " with error :" + errOpen.Error())
	}
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		logger.LogLn("error:", err)
	}
	confFileExtensionMut.Unlock()
	logger.Log("File extensions supported : " + strings.Join(configuration.Extensions, ","))

	return configuration
}

func LoadConfigurationAtOnce() FileExtension {
	return LoadConfiguration("extension-file.json")
}
