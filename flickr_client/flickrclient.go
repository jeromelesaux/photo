package flickr_client

import (
	"encoding/json"
	logger "github.com/Sirupsen/logrus"
	"github.com/jeromelesaux/photo/exifhandler"
	"github.com/jeromelesaux/photo/flickr.v2"
	"github.com/jeromelesaux/photo/modele"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photosets"
	"os"
	"sync"
	"time"
)

var currentFlickrClient *Flickr
var flickrClientOnce sync.Once
var configFlickrLock sync.Mutex
var configurationFilename = "flickr-conf.json"

func (f *Flickr) saveConfiguration() {
	configFlickrLock.Lock()
	defer configFlickrLock.Unlock()

	conf, err := os.Create(configurationFilename)
	if err != nil {
		logger.Error("Error while saving the google configuration file with error : " + err.Error())
		return
	}
	defer conf.Close()
	if err = json.NewEncoder(conf).Encode(f); err != nil {
		logger.Error("Error while unmarshaling google configuration file with error : " + err.Error())
		return
	}
}

func (f *Flickr) LoadConfiguration() {
	configFlickrLock.Lock()
	defer configFlickrLock.Unlock()

	conf, err := os.Open(configurationFilename)
	if err != nil {
		logger.Error("Error while saving the google configuration file with error : " + err.Error())
		return
	}
	defer conf.Close()
	if err = json.NewDecoder(conf).Decode(f); err != nil {
		logger.Error("Error while unmarshaling google configuration file with error : " + err.Error())
		return
	}
}

func GetCurrentFlickrClient() *Flickr {

	flickrClientOnce.Do(func() {
		currentFlickrClient = &Flickr{}
	})
	return currentFlickrClient
}

func SaveCurrentFlickrClient(f *Flickr) {
	currentFlickrClient.ApiSecret = f.ApiSecret
	currentFlickrClient.ApiKey = f.ApiKey
	currentFlickrClient.UrlAuthorization = f.UrlAuthorization
	currentFlickrClient.FlickrToken = f.FlickrToken
	currentFlickrClient.saveConfiguration()
}

type Flickr struct {
	ApiKey           string               `json:"api_key"`
	ApiSecret        string               `json:"api_secret"`
	FlickrToken      string               `json:"flickr_token,omitempty"`
	Client           *flickr.FlickrClient `json:"-"`
	RequestTok       *flickr.RequestToken `json:"-"`
	UrlAuthorization string               `json:"url_authorization,omitempty"`
}

func NewFlickrClient(apikey, apisecret string) *Flickr {
	return &Flickr{
		ApiKey:    apikey,
		ApiSecret: apisecret,
		Client:    flickr.NewFlickrClient(apikey, apisecret),
	}
}

func (f *Flickr) Connect() {
	f.Client = flickr.NewFlickrClient(f.ApiKey, f.ApiSecret)
}

func (f *Flickr) GetUrlRequestToken() string {
	requestTok, err := flickr.GetRequestToken(f.Client)
	if err != nil {
		logger.Errorf("Error while getting  request token url from flickr with error : %v", err)
	}

	// build the authorizatin URL
	f.UrlAuthorization, err = flickr.GetAuthorizeUrl(f.Client, requestTok)
	f.RequestTok = requestTok
	if err != nil {
		logger.Errorf("Error while getting authorize url from flickr with error : %v", err)
	}
	logger.Infof("flickr token url : %v", f.UrlAuthorization)

	return f.UrlAuthorization
}

func (f *Flickr) GetData(flickrChan chan *modele.PhotoResponse) {
	starttime := time.Now()
	defer func() {
		close(flickrChan)
		logger.Infof("flickr import ended in %.2f seconds .", time.Now().Sub(starttime).Seconds())
	}()
	if f.FlickrToken == "" {
		logger.Errorf("flicker token is empty import skipped.")
		flickrChan <- &modele.PhotoResponse{}
		return
	}

	accessTok, err := flickr.GetAccessToken(f.Client, f.RequestTok, f.FlickrToken)
	if err != nil {
		logger.Errorf("error while getting access token from flickr with error %v", err)
		flickrChan <- &modele.PhotoResponse{}
		return
	}
	f.Client.OAuthToken = accessTok.OAuthToken
	f.Client.OAuthTokenSecret = accessTok.OAuthTokenSecret
	photosetsResponse, err := photosets.GetList(f.Client, true, f.Client.Id, 1)
	if err != nil {
		logger.Errorf("error while getting photosets list from flickr with error %v", err)
		flickrChan <- &modele.PhotoResponse{}
		return
	}
	for _, photoset := range photosetsResponse.Photosets.Items {
		logger.Infof("Getting the flickr album %s and title %s", photoset.Id, photoset.Title)
		response := &modele.PhotoResponse{
			MachineId: modele.ORIGIN_FLICKR,
			Version:   modele.VERSION,
			Origin:    photoset.Title,
			Photos:    make([]*modele.PhotoInformations, 0),
		}
		photoListResponse, err := photosets.GetPhotos(f.Client, true, photoset.Id, photoset.Owner, 1)
		if err != nil {
			logger.Errorf("Error while getting flickr photos from photoset %s with error %v", photoset.Id, err)
		} else {
			logger.Infof("Flickr album %s get %d photos.", photoset.Id, len(photoListResponse.Photoset.Photos))
			for _, photo := range photoListResponse.Photoset.Photos {

				p := modele.NewPhotoInformations()
				p.Filename = photo.Title
				p.Md5Sum = photo.Id
				photoInfoResponse, err := flickr_v2.GetInfo(f.Client, photo.Id, f.ApiSecret)
				if err != nil {
					logger.Errorf("Error while getting photo information %s with error %v", photo.Id, err)
				} else {
					p.Thumbnail, p.Filepath = f.GetThumbnailAndOriginal(photo.Id)
					logger.Infof("Get the picture %s", p.Filepath)
				}
				exifs := f.GetExif(photoInfoResponse.Photo)
				for _, exif := range exifs {
					p.Tags[exif.Label] = exif.Raw
				}
				response.Photos = append(response.Photos, p)
			}
		}
		flickrChan <- response

	}

	return
}

func (f *Flickr) GetThumbnailAndOriginal(id string) (string, string) {
	var originalUrl string
	var thumbnail string
	response, err := flickr_v2.GetSizes(f.Client, id)
	if err != nil {
		logger.Errorf("Error while getting thumbnail with error %v for id photo %s", err, id)
		return thumbnail, originalUrl
	}
	for _, size := range response.Sizes.Sizes {
		if size.Label == "Thumbnail" {
			thumbnail, err = exifhandler.GetBase64ThumbnailUrl(size.Source)
			if err != nil {
				logger.Errorf("Error while transform thumbnail from url %s into base64 string with error %v", size.Source, err)
			}

		} else {
			if size.Label == "Original" {
				originalUrl = size.Source
			}
		}

	}
	return thumbnail, originalUrl
}

func (f *Flickr) GetExif(pinfo flickr_v2.PhotoInfo) []flickr_v2.Exif {
	response, err := flickr_v2.GetExifs(f.Client, pinfo.Id, pinfo.Secret)
	if err != nil {
		logger.Errorf("Error while getting flickr exif information on photo id %s with error %v", pinfo.Id, err)
		return make([]flickr_v2.Exif, 0)
	}
	return response.PhotoExif.Exifs
}
