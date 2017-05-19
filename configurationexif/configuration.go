// package managing the images files extensions to scan
//
package configurationexif

import (
	"encoding/json"
	logger "github.com/Sirupsen/logrus"
	"os"
	"strings"
	"sync"
)

var confFileExtensionMut sync.Mutex

// list of the file extensions available
type FileExtension struct {
	Extensions []string
}

// function loads the list of the file extensions available from the file path (configurationFile)
func LoadConfiguration(configurationFile string) FileExtension {
	configuration := FileExtension{}
	confFileExtensionMut.Lock()
	file, errOpen := os.Open(configurationFile)
	if errOpen != nil {
		logger.Error("Error while opening file " + configurationFile + " with error :" + errOpen.Error())
	}
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		logger.Error("error:", err)
	}
	confFileExtensionMut.Unlock()
	logger.Info("File extensions supported : " + strings.Join(configuration.Extensions, ","))

	return configuration
}

// function loads the structure from the static file path "extension-file.json"
func LoadConfigurationAtOnce() FileExtension {
	return LoadConfiguration("extension-file.json")
}
