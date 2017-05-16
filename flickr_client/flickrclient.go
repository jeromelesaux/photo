package flickr_client

import (
	logger "github.com/Sirupsen/logrus"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photos"
	"gopkg.in/masci/flickr.v2/photosets"
	"photo/modele"
	"sync"
)

var currentFlickrClient *Flickr
var flickrClientOnce sync.Once

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
}

type Flickr struct {
	ApiKey           string               `json:"api_key"`
	ApiSecret        string               `json:"api_secret"`
	FlickrToken      string               `json:"flickr_token"`
	Client           *flickr.FlickrClient `json:"_"`
	RequestTok       *flickr.RequestToken `json:"_"`
	UrlAuthorization string               `json:"url_authorization"`
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
	requestTok, _ := flickr.GetRequestToken(f.Client)

	// build the authorizatin URL
	f.UrlAuthorization, _ = flickr.GetAuthorizeUrl(f.Client, requestTok)
	logger.Infof("flickr token url : %v", f.UrlAuthorization)

	return f.UrlAuthorization
}

func (f *Flickr) GetData() []*modele.PhotoResponse {
	responses := make([]*modele.PhotoResponse, 0)
	accessTok, err := flickr.GetAccessToken(f.Client, f.RequestTok, f.FlickrToken)
	if err != nil {
		logger.Errorf("error while getting access token from flickr with error %v", err)
		return responses
	}
	f.Client.OAuthToken = accessTok.OAuthToken
	f.Client.OAuthTokenSecret = accessTok.OAuthTokenSecret
	photosetsResponse, err := photosets.GetList(f.Client, true, f.Client.Id, 1)
	if err != nil {
		logger.Errorf("error while getting photosets list from flickr with error %v", err)
		return responses
	}
	for _, photoset := range photosetsResponse.Photosets.Items {
		logger.Infof("Getting the flickr album %s", photoset.Id)
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
				photoInfoResponse, err := photos.GetInfo(f.Client, photo.Id, f.ApiSecret)
				if err != nil {
					logger.Errorf("Error while getting photo information %d with error %v", photo.Id, err)
				} else {
					if len(photoInfoResponse.Photo.Urls) > 0 {
						p.Filepath = photoInfoResponse.Photo.Urls[0].Url
					}
				}
				exifs := f.GetExif(photoInfoResponse.Photo)
				for _, exif := range exifs {
					p.Tags[exif.Label] = exif.Raw
				}
				response.Photos = append(response.Photos, p)
			}
		}
		responses = append(responses, response)

	}
	return responses
}

func (f *Flickr) GetExif(pinfo photos.PhotoInfo) []photos.Exif {
	response, err := photos.GetExifs(f.Client, pinfo.Id, pinfo.Secret)
	if err != nil {
		logger.Errorf("Error while getting flickr exif information on photo id %s with error %v", pinfo.Id, err)
		return make([]photos.Exif, 0)
	}
	return response.PhotoExif.Exifs
}
