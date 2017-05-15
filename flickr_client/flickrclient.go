package flickr_client

import (
	logger "github.com/Sirupsen/logrus"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photos"
	"gopkg.in/masci/flickr.v2/photosets"
	"photo/modele"
)

type Flickr struct {
	ApiKey           string               `json:"api_key"`
	ApiSecret        string               `json:"api_secret"`
	FlickrToken      string               `json:"flickr_token"`
	client           *flickr.FlickrClient `json:"_"`
	requestTok       *flickr.RequestToken `json:"_"`
	UrlAuthorization string               `json:"url_authorization"`
}

func NewFlickrClient(apikey, apisecret string) *Flickr {
	return &Flickr{
		ApiKey:    apikey,
		ApiSecret: apisecret,
		client:    flickr.NewFlickrClient(apikey, apisecret),
	}
}

func (f *Flickr) Connect() {
	f.client = flickr.NewFlickrClient(f.ApiKey, f.ApiSecret)
}

func (f *Flickr) GetUrlRequestToken() string {
	requestTok, _ := flickr.GetRequestToken(f.client)

	// build the authorizatin URL
	f.UrlAuthorization, _ = flickr.GetAuthorizeUrl(f.client, requestTok)
	logger.Infof("flickr token url : %v", f.UrlAuthorization)

	return f.UrlAuthorization
}

func (f *Flickr) GetData() []*modele.PhotoResponse {
	responses := make([]*modele.PhotoResponse, 0)
	accessTok, err := flickr.GetAccessToken(f.client, f.requestTok, f.FlickrToken)
	if err != nil {
		logger.Errorf("error while getting access token from flickr with error %v", err)
		return responses
	}
	f.client.OAuthToken = accessTok.OAuthToken
	f.client.OAuthTokenSecret = accessTok.OAuthTokenSecret
	photosetsResponse, err := photosets.GetList(f.client, true, f.client.Id, 1)
	if err != nil {
		logger.Errorf("error while getting photosets list from flickr with error %v", err)
		return responses
	}
	for _, photoset := range photosetsResponse.Photosets.Items {
		response := &modele.PhotoResponse{
			MachineId: modele.ORIGIN_FLICKR,
			Version:   modele.VERSION,
			Origin:    photoset.Title,
			Photos:    make([]*modele.PhotoInformations, 0),
		}
		photoListResponse, err := photosets.GetPhotos(f.client, true, photoset.Id, photoset.Owner, 1)
		if err != nil {
			logger.Errorf("Error while getting flickr photos from photoset %s with error %v", photoset.Id, err)
		} else {
			for _, photo := range photoListResponse.Photoset.Photos {
				p := modele.NewPhotoInformations()
				p.Filename = photo.Title
				p.Md5Sum = photo.Id
				photoInfoResponse, err := photos.GetInfo(f.client, photo.Id, f.ApiSecret)
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
	response, err := photos.GetExifs(f.client, pinfo.Id, pinfo.Secret)
	if err != nil {
		logger.Errorf("Error while getting flickr exif information on photo id %s with error %v", pinfo.Id, err)
		return make([]photos.Exif, 0)
	}
	return response.PhotoExif.Exifs
}
