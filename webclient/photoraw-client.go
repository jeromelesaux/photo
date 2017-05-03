package webclient

import (
	"encoding/json"
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"net/http"
	"photo/database"
	"photo/modele"
	"photo/slavehandler"
	"sync"
	"time"
)

type RawPhotoClient struct {
	rawPhotoChan chan string
	Album        *database.DatabaseAlbumRecord
}

func NewRawPhotoClient(albumRecord *database.DatabaseAlbumRecord) *RawPhotoClient {
	return &RawPhotoClient{
		rawPhotoChan: make(chan string, 4),
		Album:        albumRecord,
	}
}

func (p *RawPhotoClient) GetRemoteRawPhotosAlbum() []string {
	var startTime time.Time
	photosContent := make([]string, 0)

	defer func() {
		endTime := time.Now()
		computeDuration := endTime.Sub(startTime)
		logger.Infof("Job done in %f seconds\n", computeDuration.Minutes())

	}()

	// rawPhotoContent channel consumer
	wgp := &sync.WaitGroup{}
	wgp.Add(1)
	go func() {
		defer wgp.Done()
		for content := range p.rawPhotoChan {
			photosContent = append(photosContent, content)
		}
	}()
	startTime = time.Now()
	logger.Info("start generating album pdf. ")

	wg := &sync.WaitGroup{}
	for _, record := range p.Album.Records {
		logger.Info("call get photo content for " + record.Filepath + " at the machine " + record.MachineId)
		wg.Add(1)
		go p.CallGetRawPhoto(record.MachineId, record.Filepath, wg)
	}

	wg.Wait()
	close(p.rawPhotoChan)
	wgp.Wait()
	return photosContent
}

func (p *RawPhotoClient) CallGetRawPhoto(machineid, remotePath string, wg *sync.WaitGroup) {
	var startTime time.Time
	defer func() {
		endTime := time.Now()
		computeDuration := endTime.Sub(startTime)
		logger.Infof("Job done in %f seconds\n", computeDuration.Minutes())
		wg.Done()
	}()
	startTime = time.Now()
	slaves := slavehandler.GetSlaves()
	slave := slaves.Slaves[machineid]
	if slave == nil {
		logger.Error("Slave for machineId:" + machineid + " not found. Skiped")
		p.rawPhotoChan <- ""
		return
	}
	client := &http.Client{}
	uri := fmt.Sprintf("%s:%d/photo?filepath=%s", slave.Url, slave.Port, remotePath)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- ""
		return
	}
	logger.Info("Calling uri : " + uri)
	response, err := client.Do(request)
	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- ""
		return
	}
	content := &modele.RawPhoto{}

	if err := json.NewDecoder(response.Body).Decode(content); err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- content.Data
		return
	}
	p.rawPhotoChan <- content.Data
	return

}
