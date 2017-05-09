package configurationapp

import (
	"encoding/json"
	logger "github.com/Sirupsen/logrus"
	"os"
	"sync"
)

type Configuration struct {
	DatabasePath string `json:"database_path"`
	GoogleID     string `json:"google_id"`
	GoogleUser   string `json:"google_user"`
	GoogleSecret string `json:"google_secret"`
}

var confPhotoExifMut sync.Mutex
var configuration *Configuration
var configurationFilepath string

func LoadPhotoExifConfiguration(configurationPathfile string) *Configuration {
	configurationFilepath = configurationPathfile
	confPhotoExifMut.Lock()
	file, errOpen := os.Open(configurationPathfile)
	if errOpen != nil {
		logger.Error("Error while opening file " + configurationPathfile + " with error :" + errOpen.Error())
	}
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		logger.Error("error:", err)
	}
	logger.Debug(*configuration)
	confPhotoExifMut.Unlock()
	return configuration
}

func GetConfiguration() *Configuration {
	return configuration
}

func (c *Configuration) Save() error {
	confPhotoExifMut.Lock()
	defer confPhotoExifMut.Unlock()
	file, errOpen := os.Create(configurationFilepath)
	if errOpen != nil {
		logger.Error("Error while opening file " + configurationFilepath + " with error :" + errOpen.Error())
	}
	err := json.NewEncoder(file).Encode(c)
	if err != nil {
		logger.Error("error:", err)
	}
	return nil
}
