package routes

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"photo/client"
	"photo/exifhandler"
	"photo/folder"
	"photo/logger"
	"photo/modele"
	"strconv"
	"time"
)

func Browse(w http.ResponseWriter, r *http.Request) {
	//usr, _ := user.Current()
	starttime := time.Now()
	directorypath := r.URL.Query().Get("value")
	response := &modele.DirectoryItemResponse{
		Name:             "Root",
		Path:             directorypath,
		JstreeAttributes: modele.NewJSTreeAttribute(),
		Directories:      make([]*modele.DirectoryItemResponse, 0),
	}
	if directorypath == "" {
		JsonAsResponse(w, response)
		return
	} else {
		if directorypath[len(directorypath)-1] != '/' {
			directorypath += "/"
		}
	}

	err := filepath.Walk(directorypath, folder.ScanDirectory(response))
	if err != nil && err.Error() != "" {
		response.Message = err.Error()
		logger.Log(err.Error())
	}
	logger.Log("Scan directory completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	JsonAsResponse(w, response)
}

func ScanFolders(w http.ResponseWriter, r *http.Request) {
	conf := modele.GetConfiguration()
	folders := &modele.FolderToScan{}
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}

	err := json.NewDecoder(r.Body).Decode(folders)
	logger.LogLn(&folders)
	var response string
	if err != nil {
		response = err.Error()
	} else {
		response = "Scans launched."
	}
	go client.ScanFoldersClient(folders.Folders,conf)

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
	pinfos, err := exifhandler.GetPhotoInformations(filepath)
	response.Photos = append(response.Photos, pinfos)
	if err != nil {
		response.Message = err.Error()
	}
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
	pinfos, err := exifhandler.GetPhotosInformations(directorypath, modele.LoadConfigurationAtOnce())
	response.Photos = pinfos
	if err != nil {
		response.Message = err.Error()
	}
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
