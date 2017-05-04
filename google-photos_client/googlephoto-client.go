package google_photos_client

import (
	"encoding/json"
	logger "github.com/Sirupsen/logrus"
	"os"
	"sync"
)

type GooglePhotoClient struct {
	tokenCacheFile string
	UserID         string
	Secret         string
	ID             string
}

var configGoogleLock sync.Mutex
var configurationFilename = "google-conf.json"

func NewGooglePhotoClient(userId string, id string, secret string) *GooglePhotoClient {
	return &GooglePhotoClient{
		ID:             id,
		UserID:         userId,
		Secret:         secret,
		tokenCacheFile: "token-cache.json",
	}
}


func NewGooglePhotoClientFromConfiguration() *GooglePhotoClient {
	g := &GooglePhotoClient{}
	configGoogleLock.Lock()
	f,err := os.Open(configurationFilename)

	defer configGoogleLock.Unlock()
	if err != nil {
		logger.Error("Error while saving the google configuration file with error : " + err.Error())
		return g
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(g); err != nil {
		logger.Error("Error while unmarshaling google configuration file with error : " + err.Error())
		return   g
	}
	return g
}

func (g *GooglePhotoClient) saveConfigutation() {
	configGoogleLock.Lock()
	defer configGoogleLock.Unlock()

	f, err := os.Create(configurationFilename)
	if err != nil {
		logger.Error("Error while saving the google configuration file with error : " + err.Error())
		return
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(g); err != nil {
		logger.Error("Error while unmarshaling google configuration file with error : " + err.Error())
		return
	}

}




