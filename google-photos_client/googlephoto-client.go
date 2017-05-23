package google_photos_client

import (
	"encoding/json"
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"github.com/tgulacsi/picago"
	"net/http"
	"os"
	"photo/exifhandler"
	"photo/modele"
	"strconv"
	"sync"
)

type GooglePhotoClient struct {
	client *http.Client
	UserID string `json:"user_id"`
	Secret string `json:"secret"`
	ID     string `json:"id"`
}

var configGoogleLock sync.Mutex
var configurationFilename = "google-conf.json"
var tokenCacheFile = "token-cache.json"

func NewGooglePhotoClient(userId string, id string, secret string) *GooglePhotoClient {
	return &GooglePhotoClient{
		ID:     id,
		UserID: userId,
		Secret: secret,
	}
}

func NewGooglePhotoClientFromConfiguration() *GooglePhotoClient {
	g := &GooglePhotoClient{}
	configGoogleLock.Lock()
	f, err := os.Open(configurationFilename)

	defer configGoogleLock.Unlock()
	if err != nil {
		logger.Error("Error while saving the google configuration file with error : " + err.Error())
		return g
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(g); err != nil {
		logger.Error("Error while unmarshaling google configuration file with error : " + err.Error())
		return g
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

func (g *GooglePhotoClient) Connect() error {
	client, err := picago.NewClient(g.ID, g.Secret, "", tokenCacheFile)
	if err != nil {
		logger.Error("Error while connecting to google photo with error : " + err.Error())
		return err
	} else {
		g.client = client
		return nil
	}
}

func (g *GooglePhotoClient) GetData() []*modele.PhotoResponse {
	responses := make([]*modele.PhotoResponse, 0)

	albums, err := picago.GetAlbums(g.client, g.UserID)
	if err != nil {
		logger.Errorf("error listing albums: %v", err)
	}
	logger.Infof("user %s has %d albums.", g.UserID, len(albums))
	for _, album := range albums {
		response := &modele.PhotoResponse{
			MachineId: modele.ORIGIN_GOOGLE,
			Version:   modele.VERSION,
			Origin:    album.Name,
			Photos:    make([]*modele.PhotoInformations, 0),
		}
		photos, err := picago.GetPhotos(g.client, g.UserID, album.ID)
		if err != nil {
			logger.Error("error with error message :" + err.Error())
			continue
		}
		logger.Info(album.Name + " contains " + strconv.Itoa(len(photos)) + " photos.")

		for _, photo := range photos {
			p := modele.NewPhotoInformations()
			p.Filename = photo.Filename
			p.Md5Sum = photo.Exif.UID
			if photo.Exif != nil {
				p.Tags["exposure"] = fmt.Sprintf("%f", photo.Exif.Exposure)
				p.Tags["flash"] = fmt.Sprintf("%t", photo.Exif.Flash)
				p.Tags["focal length"] = fmt.Sprintf("%f", photo.Exif.FocalLength)
				p.Tags["fstop"] = fmt.Sprintf("%f", photo.Exif.FStop)
				p.Tags["iso"] = fmt.Sprintf("%d", photo.Exif.ISO)
				p.Tags["make"] = photo.Exif.Make
				p.Tags["model"] = photo.Exif.Model
				p.Tags["timestamp"] = fmt.Sprintf("%d", photo.Exif.Timestamp)
				p.Tags["uid"] = photo.Exif.UID
			}
			p.Filepath = photo.URL
			p.Thumbnail, _ = exifhandler.GetBase64ThumbnailUrl(photo.URL)
			p.Tags["with"] = fmt.Sprintf("%d", photo.Width)
			p.Tags["height"] = fmt.Sprintf("%d", photo.Height)
			p.Tags["location"] = photo.Location
			p.Tags["description"] = photo.Description
			p.Tags["latitude"] = fmt.Sprintf("%.2f", photo.Latitude)
			p.Tags["longitude"] = fmt.Sprintf("%.2f", photo.Longitude)
			response.Photos = append(response.Photos, p)

		}
		responses = append(responses, response)
	}
	return responses
}
