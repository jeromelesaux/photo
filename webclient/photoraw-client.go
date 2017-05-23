package webclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"photo/database"
	"photo/exifhandler"
	"photo/modele"
	"photo/slavehandler"
	"strings"
	"sync"
	"time"
)

type RawPhotoClient struct {
	rawPhotoChan chan *modele.ExportPdf
	Album        *database.DatabaseAlbumRecord
}

func NewRawPhotoClient(albumRecord *database.DatabaseAlbumRecord) *RawPhotoClient {
	return &RawPhotoClient{
		rawPhotoChan: make(chan *modele.ExportPdf, 4),
		Album:        albumRecord,
	}
}

func (p *RawPhotoClient) GetRemoteRawPhotosAlbum() []*modele.ExportPdf {
	var startTime time.Time
	photosFilenames := make([]*modele.ExportPdf, 0)

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
			go p.CallGetRemoteRawPhoto(record.Filepath, wg, true)
		case modele.ORIGIN_FLICKR:
			go p.CallGetRemoteRawPhoto(record.Filepath, wg, true)
		default:
			go p.CallGetRawPhoto(record.MachineId, record.Filepath, wg, true)
		}
	}

	wg.Wait()
	close(p.rawPhotoChan)
	wgp.Wait()
	return photosFilenames
}

func (p *RawPhotoClient) CallGetRemoteRawPhoto(remotePath string, wg *sync.WaitGroup, returnFilenameSaved bool) {
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
		p.rawPhotoChan <- &modele.ExportPdf{}
		return
	}
	logger.Info("Calling uri : " + remotePath)
	response, err := client.Do(request)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- &modele.ExportPdf{}
		return
	}
	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()
	img, _, err := image.Decode(response.Body)
	if err != nil {
		logger.Error("error with : " + err.Error() + " for uri:" + remotePath)
		p.rawPhotoChan <- &modele.ExportPdf{}
		return
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		logger.Errorf("Cannot convert to png file : %v", err)
	}

	if returnFilenameSaved {
		rand.Seed(time.Now().UTC().UnixNano())
		filename := fmt.Sprintf("img_%d.jpg", rand.Int())
		f, err := os.Create(filename)
		if err != nil {
			logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
			p.rawPhotoChan <- &modele.ExportPdf{}
		} else {
			defer f.Close()
			if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 99}); err != nil {
				logger.Infof("error in encoding temporary file %s with error %v", filename, err.Error())
			}

			e := &modele.ExportPdf{Filename: filename,
				Orientation: exifhandler.Orientation(img),
			}
			p.rawPhotoChan <- e
		}
	} else {

		e := &modele.ExportPdf{Base64Content: base64.StdEncoding.EncodeToString(buf.Bytes()),
			Orientation: exifhandler.Orientation(img),
		}
		p.rawPhotoChan <- e
	}
	return

}

func (p *RawPhotoClient) CallGetRawPhoto(machineid, remotePath string, wg *sync.WaitGroup, returnFilenameSaved bool) {
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
		p.rawPhotoChan <- &modele.ExportPdf{}
		return
	}
	client := &http.Client{}
	uri := fmt.Sprintf("%s:%d/photo?filepath=%s", slave.Url, slave.Port, strings.Replace(remotePath, " ", "%20", -1))
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- &modele.ExportPdf{}
		return
	}
	logger.Info("Calling uri : " + uri)
	response, err := client.Do(request)
	if err != nil {
		logger.Error("error with : " + err.Error())
		p.rawPhotoChan <- &modele.ExportPdf{}
		return
	}
	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()
	content := &modele.RawPhoto{}
	logger.Info(response.Body)
	if err := json.NewDecoder(response.Body).Decode(content); err != nil {
		logger.Error("error with : " + err.Error() + " for uri:" + uri)
		p.rawPhotoChan <- &modele.ExportPdf{}
		return
	}
	if returnFilenameSaved {
		rand.Seed(time.Now().UTC().UnixNano())
		filename := fmt.Sprintf("img_%d.jpg", rand.Int())
		f, err := os.Create(filename)
		if err != nil {
			logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
		} else {
			defer f.Close()
			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(content.Data))
			img, err := png.Decode(reader)
			if err != nil {
				logger.Infof("error in creating temporary file %s with error %v", filename, err.Error())
				os.Remove(filename)
				p.rawPhotoChan <- &modele.ExportPdf{}
			} else {
				if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 99}); err != nil {
					logger.Infof("error in encoding temporary file %s with error %v", filename, err.Error())
				}
				e := &modele.ExportPdf{Filename: filename,
					Orientation: exifhandler.Orientation(img),
				}
				p.rawPhotoChan <- e
			}
		}

	} else {
		e := &modele.ExportPdf{Base64Content: content.Data,
			Orientation: content.Orientation,
		}
		p.rawPhotoChan <- e
	}
	return

}
