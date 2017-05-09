package configurationexif

import (
	"encoding/json"
	logger "github.com/Sirupsen/logrus"
	"os"
	"strings"
	"sync"
)

var confFileExtensionMut sync.Mutex

type FileExtension struct {
	Extensions []string
}

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

func LoadConfigurationAtOnce() FileExtension {
	return LoadConfiguration("extension-file.json")
}
