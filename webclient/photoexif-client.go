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

	"photo/configurationapp"
	"photo/configurationexif"
	"sync"
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
		logger.Infof("Job done in %f seconds\n", computeDuration.Minutes())

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
	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()
	photoResponse := &modele.PhotoResponse{}
	if err := json.NewDecoder(response.Body).Decode(photoResponse); err != nil {
		logger.Error("error with : " + err.Error())
		p.photoResponseChan <- photoResponse
		return
	}
	logger.Info("Found " + strconv.Itoa(len(photoResponse.Photos)) + " images in " + remotePath)
	photoResponse.MachineId = salve.Name
	p.photoResponseChan <- photoResponse
	return
}

type Job struct {
	r   string
	sid string
}

func (p *PhotoExifClient) ScanFoldersClient(remotepaths []string, slaveid string, conf *configurationapp.Configuration) {

	wgp := sync.WaitGroup{}
	wgp.Add(1)
	go func() {
		defer wgp.Done()
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

	slavesConfig := slavehandler.GetSlaves()

	if len(slavesConfig.Slaves) == 0 {
		logger.Error("No slave registered, skip action")
		return
	}

	c := make(chan Job, 100)
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for d := range c {
				logger.Info("Sending to traitment " + d.r)
				if slave := slavesConfig.Slaves[d.sid]; slave != nil {
					logger.Info("Exec search to " + slave.Name + " address " + slave.Url + " for directory " + d.r)
					p.scanExifClient(d.r, slave)
				}

			}
			logger.Info("Finished all treatments")
			modele.PostActionMessage("scan folders ended for machineid " + slaveid)
		}()
	}

	for _, remotepath := range remotepaths {
		logger.Info("Sending to traitment " + remotepath)

		j := Job{r: remotepath, sid: slaveid}
		c <- j

	}
	close(c)
	wg.Wait()

	close(p.photoResponseChan)

	wgp.Wait()

}

func (p *PhotoExifClient) GetFileExtensionValues(slave *slavehandler.Slave) (error, *configurationexif.FileExtension) {
	var startTime time.Time

	defer func() {
		endTime := time.Now()
		computeDuration := endTime.Sub(startTime)
		logger.Infof("Job done in %f seconds\n", computeDuration.Minutes())

	}()
	startTime = time.Now()
	client := &http.Client{}
	uri := fmt.Sprintf("%s:%d/getfileextension", slave.Url, slave.Port)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Error("error with : " + err.Error())
		return err, &configurationexif.FileExtension{}
	}
	logger.Info("Calling uri : " + uri)
	response, err := client.Do(request)
	if err != nil {
		logger.Error("error with : " + err.Error())
		return err, &configurationexif.FileExtension{}
	}
	defer response.Body.Close()
	extensions := &configurationexif.FileExtension{}
	if err := json.NewDecoder(response.Body).Decode(extensions); err != nil {
		logger.Error("error with : " + err.Error())
		return nil, &configurationexif.FileExtension{}
	}
	return nil, extensions
}

func (p *PhotoExifClient) GetThumbnails(responses []*database.DatabasePhotoRecord, size string) []*database.DatabasePhotoRecord {
	slavesConfig := slavehandler.GetSlaves()
	finalResponses := make([]*database.DatabasePhotoRecord, 0)
	for _, response := range responses {

		if ok := slavesConfig.Slaves[response.MachineId]; ok == nil {
			logger.Warn("No client found for machine id " + response.MachineId)
		} else {
			err, data := p.GetThumbnail(ok, response.Filepath)
			if err != nil {
				logger.Error("error while getting thumbnail from machine " + ok.Name + " with error " + err.Error())
			} else {
				//logger.Infof("filesize :%d", len(data))
				switch size {
				case modele.FILESIZE_LITTLE:
					response.Image = data
					finalResponses = append(finalResponses, response)
				case modele.FILESIZE_MEDIUM:
					if len(data) > 15000 {
						response.Image = data
						finalResponses = append(finalResponses, response)
					}
				case modele.FILESIZE_BIG:
					if len(data) > 25000 {
						response.Image = data
						finalResponses = append(finalResponses, response)
					}
				default:
					response.Image = data
					finalResponses = append(finalResponses, response)
				}
			}
		}
	}

	return finalResponses
}

func (p *PhotoExifClient) GetThumbnail(slave *slavehandler.Slave, path string) (error, string) {
	var startTime time.Time
	var image string
	defer func() {
		endTime := time.Now()
		computeDuration := endTime.Sub(startTime)
		logger.Infof("Job done in %f seconds\n", computeDuration.Minutes())

	}()
	startTime = time.Now()
	client := &http.Client{}
	uri := fmt.Sprintf("%s:%d/thumbnail?filepath=%s", slave.Url, slave.Port, path)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Error("error with : " + err.Error())
		return err, ""
	}
	logger.Info("Calling uri : " + uri)
	response, err := client.Do(request)
	if err != nil {
		logger.Error("error with : " + err.Error())
		return err, ""
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&image); err != nil {
		logger.Error("error with : " + err.Error())
		return nil, ""
	}
	return nil, image
}
