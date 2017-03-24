package routes

import (
	"encoding/json"
	"net/http"
	"os/user"
	"path/filepath"
	"photo/exifhandler"
	"photo/folder"
	"photo/logger"
	"photo/modele"
	"strconv"
	"time"
)

func Browse(w http.ResponseWriter, r *http.Request) {
	usr, _ := user.Current()
	starttime := time.Now()
	directorypath := r.URL.Query().Get("value")
	if directorypath == "" {
		directorypath = usr.HomeDir
	} else {
		if directorypath[len(directorypath)-1] != '/' {
			directorypath += "/"
		}
	}

	response := &modele.DirectoryItemResponse{
		Name:             "Root",
		Path:             "browse?value=" + directorypath,
		JstreeAttributes: modele.NewJSTreeAttribute(),
		Directories:      make([]*modele.DirectoryItemResponse, 0),
	}
	filepath.Walk(directorypath, folder.ScanDirectory(response))
	logger.Log("Scan directory completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	JsonAsResponse(w, response)
}

func GetFileInformations(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	filepath := r.URL.Query().Get("value")
	logger.Log("file to scan " + filepath)
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.TagsPhoto, 0),
	}
	response.Photos = append(response.Photos, exifhandler.GetPhotoInformations(filepath))
	logger.Log("Scan completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	logger.Log(strconv.Itoa(len(response.Photos)) + " images found")
	JsonAsResponse(w, response)
}

func GetDirectoryInformations(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.TagsPhoto, 0),
	}
	directorypath := r.URL.Query().Get("value")
	logger.Log("directory to scan " + directorypath)
	response.Photos = exifhandler.GetPhotosInformations(directorypath, modele.LoadConfigurationAtOnce())
	logger.Log("Scan completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	logger.Log(strconv.Itoa(len(response.Photos)) + " images found")
	JsonAsResponse(w, response)
}

func JsonAsResponse(w http.ResponseWriter, o interface{}) {
	js, err := json.Marshal(o)
	if err != nil {
		logger.Log("Error while marshalling  object")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application-json")
	w.Write(js)
}
