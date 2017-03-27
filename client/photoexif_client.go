package client

import (
	"photo/logger"
	"photo/modele"

	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//var traitmentChan = make(chan int, 10)
var photoResponseChan = make(chan *modele.PhotoResponse)

func scanExifClient(remotePath string, conf *modele.Configuration, wg *sync.WaitGroup) {
	var startTime time.Time

	defer func() {
		endTime := time.Now()
		computeDuration := endTime.Sub(startTime)
		logger.Logf("Job done in %f seconds\n", computeDuration.Minutes())
		wg.Done()
		//<-traitmentChan
	}()
	startTime = time.Now()
	logger.Log(remotePath + " started to scan ")
	client := &http.Client{}
	uri := fmt.Sprintf("%s:%d%s?value=%s", conf.PhotoExifUrl, conf.PhotoExifPort, conf.PhotoExifAction, remotePath)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Log("error with : " + err.Error())
		photoResponseChan <- &modele.PhotoResponse{}
		return
	}
	logger.Log("Calling uri : " + uri)
	response, err := client.Do(request)
	if err != nil {
		logger.Log("error with : " + err.Error())
		photoResponseChan <- &modele.PhotoResponse{}
		return
	}
	defer response.Body.Close()
	photoResponse := &modele.PhotoResponse{}
	if err := json.NewDecoder(response.Body).Decode(photoResponse); err != nil {
		logger.Log("error with : " + err.Error())
		photoResponseChan <- photoResponse
		return
	}
	logger.Log("Found " + strconv.Itoa(len(photoResponse.Photos)) + " images in " + remotePath)

	photoResponseChan <- photoResponse
	return
}

func ScanFoldersClient(remotepaths []string, conf *modele.Configuration) {
	wg := new(sync.WaitGroup)
	for _, remotepath := range remotepaths {
		logger.Log("Sending to traitment " + remotepath)
		//traitmentChan <- 1
		wg.Add(1)
		go scanExifClient(remotepath, conf, wg)
	}

	go func() {

		for pr := range photoResponseChan {
			logger.Log("message received")
			logger.LogLn(*pr)
		}
	}()
	go func() {
		wg.Wait()
		//close(photoResponseChan)
		logger.Log("Traitment ended")
		return
	}()

}
