package modele

import (
	"encoding/json"
	"fmt"
	"os"
	"photo/logger"
	"strings"
	"sync"
)

type TagsPhoto struct {
	Tags     map[string]string `json:"tags"`
	Md5Sum   string            `json:"md5sum"`
	Filename string            `json:"filename"`
	Filepath string            `json:"filepath"`
}

type PhotoResponse struct {
	Version string       `json:"version"`
	Photos  []*TagsPhoto `json:"photos"`
}

type FileExtension struct {
	Extensions []string
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
	Name             string                   `json:"text"`
	Path             string                   `json:"id"`
	Directories      []*DirectoryItemResponse `json:"children"`
	Parent           *DirectoryItemResponse   `json:"-"`
	Deep             int                      `json:"-"`
	JstreeAttributes *JSTreeAttribute         `json:"state"`
}

const VERSION = "1.0Beta"

var mut sync.Mutex

func LoadConfiguration(configurationFile string) FileExtension {
	configuration := FileExtension{}
	mut.Lock()
	file, errOpen := os.Open(configurationFile)
	if errOpen != nil {
		logger.Log("Error while opening file " + configurationFile + " with error :" + errOpen.Error())
	}
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	mut.Unlock()
	logger.Log("File extensions supported : " + strings.Join(configuration.Extensions, ","))

	return configuration
}

func LoadConfigurationAtOnce() FileExtension {
	return LoadConfiguration("extension-file.json")
}
