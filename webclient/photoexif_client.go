package webclient

import (
	"photo/logger"
	"photo/modele"

	"encoding/json"
	"fmt"
	"net/http"
	"photo/database"
	"photo/slavehandler"
	"strconv"
	"sync"
	"time"
)

//var traitmentChan = make(chan int, 10)
var photoResponseChan = make(chan *modele.PhotoResponse)

func scanExifClient(remotePath string, salve *slavehandler.Slave, wg *sync.WaitGroup) {
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
	uri := fmt.Sprintf("%s:%d%s?value=%s", salve.Url, salve.Port, salve.Action, remotePath)
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
	slavesConfig := slavehandler.GetSlaves()
	if len(slavesConfig.Slaves) == 0 {
		logger.Log("No slave registered, skip action")
		return
	}
	wg := new(sync.WaitGroup)
	for _, remotepath := range remotepaths {
		logger.Log("Sending to traitment " + remotepath)
		//traitmentChan <- 1
		for _, slave := range slavesConfig.Slaves {
			wg.Add(1)
			go scanExifClient(remotepath, slave, wg)
		}
	}

	go func() {
		for pr := range photoResponseChan {
			if len(pr.Photos) > 0 {
				err := database.InsertNewData(pr)
				if err != nil {
					logger.Log("Error insert data with error" + err.Error())
				}
			}
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
