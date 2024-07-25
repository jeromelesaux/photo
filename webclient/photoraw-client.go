package webclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/jeromelesaux/photo/database"
	"github.com/jeromelesaux/photo/exifhandler"
	"github.com/jeromelesaux/photo/modele"
	"github.com/jeromelesaux/photo/slavehandler"
	logger "github.com/sirupsen/logrus"
)

type RawPhotoClient struct {
	rawPhotoChan chan *modele.ExportRawPhoto
	Album        *database.DatabaseAlbumRecord
}

func NewRawPhotoClient(albumRecord *database.DatabaseAlbumRecord) *RawPhotoClient {
	return &RawPhotoClient{
		rawPhotoChan: make(chan *modele.ExportRawPhoto, 4),
		Album:        albumRecord,
	}
}

func NewRawPhotoClientWithData(records []*database.DatabasePhotoRecord) *RawPhotoClient {
	return &RawPhotoClient{
		rawPhotoChan: make(chan *modele.ExportRawPhoto, 4),
		Album:        database.NewDataseAlbumRecordWithData(records),
	}
}

func (p *RawPhotoClient) GetRemoteRawPhotosAlbum(saveIntoAFile bool) []*modele.ExportRawPhoto {
	var startTime time.Time
	photosFilenames := make([]*modele.ExportRawPhoto, 0)

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
		for filename := range p.rawPhotoChan {
			photosFilenames = append(photosFilenames, filename)
		}
	}()
	startTime = time.Now()
	logger.Info("start generating album pdf. ")

	wg := &sync.WaitGroup{}
	for _, record := range p.Album.Records {
		logger.Info("call get photo content for " + record.Filepath + " at the machine " + record.MachineId)
		wg.Add(1)
		switch record.MachineId {
		case modele.ORIGIN_GOOGLE:
			go p.CallGetRemoteRawPhoto(record.Filepath, wg, saveIntoAFile)
		case modele.ORIGIN_FLICKR:
			go p.CallGetRemoteRawPhoto(record.Filepath, wg, saveIntoAFile)
		default:
			go p.CallGetRawPhoto(record.MachineId, record.Filepath, wg, saveIntoAFile)
		}
	}

	wg.Wait()
	close(p.rawPhotoChan)
	wgp.Wait()
	return photosFilenames
}

func (p *RawPhotoClient) CallGetRemoteRawPhoto(remotePath string, wg *sync.WaitGroup, saveIntoAFile bool) {
	var startTime time.Time
	defer func() {
		endTime := time.Now()
		computeDuration := endTime.Sub(startTime)
		logger.Infof("Job done in %f seconds\n", computeDuration.Minutes())
		wg.Done()
	}()
	startTime = time.Now()

	client := &http.Client{}
	request, err := http.NewRequest("GET", remotePath, nil)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- &modele.ExportRawPhoto{}
		return
	}
	logger.Info("Calling uri : " + remotePath)
	response, err := client.Do(request)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- &modele.ExportRawPhoto{}
		return
	}
	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()

	//_,params,err := mime.ParseMediaType(response.Header.Get("Content-Disposition"))
	//if err != nil {
	//	logger.Errorf("Cannot parse media type from header with error %v",err)
	//}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Error("error with : " + err.Error() + " for uri:" + remotePath)
		p.rawPhotoChan <- &modele.ExportRawPhoto{}
		return
	}

	if saveIntoAFile {

		readerContent := bytes.NewBuffer(content)
		img, _, err := image.Decode(readerContent)
		if err != nil {
			logger.Error("error with : " + err.Error() + " for uri:" + remotePath + " cannot decode image.")
			p.rawPhotoChan <- &modele.ExportRawPhoto{}
			return
		}

		rand.Seed(time.Now().UTC().UnixNano())
		filename := fmt.Sprintf("img_%d.jpg", rand.Int())
		f, err := os.Create(filename)
		if err != nil {
			logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
			p.rawPhotoChan <- &modele.ExportRawPhoto{}
		} else {
			defer f.Close()
			if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 99}); err != nil {
				logger.Infof("error in encoding temporary file %s with error %v", filename, err.Error())
			}

			e := &modele.ExportRawPhoto{Filename: filename,
				Orientation: exifhandler.OrientationFromImg(img),
			}
			p.rawPhotoChan <- e
		}
	} else {

		e := &modele.ExportRawPhoto{
			Filename:      path.Base(remotePath),
			Base64Content: base64.StdEncoding.EncodeToString(content),
			Orientation:   "",
		}
		p.rawPhotoChan <- e
	}
	return

}

func (p *RawPhotoClient) CallGetRawPhoto(machineid, remotePath string, wg *sync.WaitGroup, saveIntoAFile bool) {
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
		p.rawPhotoChan <- &modele.ExportRawPhoto{}
		return
	}
	client := &http.Client{}
	uri := fmt.Sprintf("%s:%d/photo?filepath=%s", slave.Url, slave.Port, strings.Replace(remotePath, " ", "%20", -1))
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- &modele.ExportRawPhoto{}
		return
	}
	logger.Info("Calling uri : " + uri)
	response, err := client.Do(request)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- &modele.ExportRawPhoto{}
		return
	}
	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()
	content := &modele.ExportRawPhoto{}
	logger.Info(response.Body)
	if err := json.NewDecoder(response.Body).Decode(content); err != nil {
		logger.Error("error with : " + err.Error() + " for uri:" + uri)
		p.rawPhotoChan <- &modele.ExportRawPhoto{}
		return
	}
	if saveIntoAFile {
		rand.Seed(time.Now().UTC().UnixNano())
		filename := fmt.Sprintf("img_%d.jpg", rand.Int())
		f, err := os.Create(filename)
		if err != nil {
			logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
		} else {
			defer f.Close()
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(content.Base64Content))
			img, err := png.Decode(reader)
			if err != nil {
				logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
				os.Remove(filename)
				p.rawPhotoChan <- &modele.ExportRawPhoto{}
			} else {
				if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 99}); err != nil {
					logger.Infof("error in encoding temporary file %s with error %v", filename, err.Error())
				}
				e := &modele.ExportRawPhoto{Filename: filename,
					Orientation: exifhandler.OrientationFromImg(img),
				}
				p.rawPhotoChan <- e
			}
		}

	} else {
		e := &modele.ExportRawPhoto{
			Filename:      path.Base(remotePath),
			Base64Content: content.Base64Content,
			Orientation:   content.Orientation,
		}
		p.rawPhotoChan <- e
	}
	return

}
