package webclient

import (
	logger "github.com/Sirupsen/logrus"
	"photo/modele"

	"encoding/json"
	"fmt"
	"net/http"
	"photo/database"
	"photo/slavehandler"
	"strconv"

	"time"
)

type PhotoExifClient struct {
	photoResponseChan chan *modele.PhotoResponse
}

func NewPhotoExifClient() *PhotoExifClient {
	return &PhotoExifClient{photoResponseChan: make(chan *modele.PhotoResponse, 5)}
}

func (p *PhotoExifClient) scanExifClient(remotePath string, salve *slavehandler.Slave) {
	var startTime time.Time

	defer func() {
		endTime := time.Now()
		computeDuration := endTime.Sub(startTime)
		logger.Info("Job done in %f seconds\n", computeDuration.Minutes())

	}()
	startTime = time.Now()
	logger.Info(remotePath + " started to scan ")
	client := &http.Client{}
	uri := fmt.Sprintf("%s:%d%s?value=%s", salve.Url, salve.Port, salve.Action, remotePath)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.photoResponseChan <- &modele.PhotoResponse{}
		return
	}
	logger.Info("Calling uri : " + uri)
	response, err := client.Do(request)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.photoResponseChan <- &modele.PhotoResponse{}
		return
	}
	defer response.Body.Close()
	photoResponse := &modele.PhotoResponse{}
	if err := json.NewDecoder(response.Body).Decode(photoResponse); err != nil {
		logger.Error("error with : " + err.Error())
		p.photoResponseChan <- photoResponse
		return
	}
	logger.Info("Found " + strconv.Itoa(len(photoResponse.Photos)) + " images in " + remotePath)

	p.photoResponseChan <- photoResponse
	return
}

func (p *PhotoExifClient) ScanFoldersClient(remotepaths []string, slaveid string, conf *modele.Configuration) {

	slavesConfig := slavehandler.GetSlaves()

	if len(slavesConfig.Slaves) == 0 {
		logger.Error("No slave registered, skip action")
		return
	}

	for _, remotepath := range remotepaths {
		logger.Info("Sending to traitment " + remotepath)
		//traitmentChan <- 1
		if slave := slavesConfig.Slaves[slaveid]; slave != nil {
			logger.Info("Exec search to " + slave.Name + " address " + slave.Url + " for directory " + remotepath)
			go p.scanExifClient(remotepath, slave)
		}
	}

	go func() {

		//var pr *modele.PhotoResponse
		for pr := range p.photoResponseChan {
			if len(pr.Photos) > 0 {
				db, err := database.NewDatabaseHandler()
				if err != nil {
					return
				}
				err = db.InsertNewData(pr)
				if err != nil {
					logger.Error("Error insert data with error" + err.Error())
				}
			}
			logger.Info("message received")
			logger.Debug(*pr)
		}

	}()
}

func (p *PhotoExifClient) GetFileExtensionValues(slave *slavehandler.Slave) (error, *modele.FileExtension) {
	var startTime time.Time

	defer func() {
		endTime := time.Now()
		computeDuration := endTime.Sub(startTime)
		logger.Info("Job done in %f seconds\n", computeDuration.Minutes())

	}()
	startTime = time.Now()
	client := &http.Client{}
	uri := fmt.Sprintf("%s:%d/getfileextension", slave.Url, slave.Port)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Error("error with : " + err.Error())
		return err, &modele.FileExtension{}
	}
	logger.Info("Calling uri : " + uri)
	response, err := client.Do(request)
	if err != nil {
		logger.Error("error with : " + err.Error())
		return err, &modele.FileExtension{}
	}
	defer response.Body.Close()
	extensions := &modele.FileExtension{}
	if err := json.NewDecoder(response.Body).Decode(extensions); err != nil {
		logger.Error("error with : " + err.Error())
		return nil, &modele.FileExtension{}
	}
	return nil, extensions
}
