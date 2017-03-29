package routes

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"photo/database"
	"photo/exifhandler"
	"photo/folder"
	"photo/logger"
	"photo/modele"
	"photo/webclient"
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
	go webclient.ScanFoldersClient(folders.Folders, conf)

	JsonAsResponse(w, response)
}

func GetFileInformations(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	filepathValue := r.URL.Query().Get("value")
	logger.Log("file to scan " + filepathValue)
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.TagsPhoto, 0),
	}
	pinfos, err := exifhandler.GetPhotoInformations(filepathValue)
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

func QueryExtension(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	filename := r.URL.Query().Get("value")
	response, err := database.QueryExtenstion(filename)
	logger.Log("QueryFilename completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	if err != nil {
		JsonAsResponse(w, err)
	} else {
		JsonAsResponse(w, response)
	}
}

func QueryExif(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	pattern := r.URL.Query().Get("value")
	exiftag := r.URL.Query().Get("exif")
	response, err := database.QueryExifTag(pattern, exiftag)
	logger.Log("QueryFilename completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	if err != nil {
		JsonAsResponse(w, err)
	} else {
		JsonAsResponse(w, response)
	}
}

func QueryFilename(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	filename := r.URL.Query().Get("value")
	response, err := database.QueryFilename(filename)
	logger.Log("QueryFilename completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	if err != nil {
		JsonAsResponse(w, err)
	} else {
		JsonAsResponse(w, response)
	}
}

func QueryAll(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	response, err := database.QueryAll()
	logger.Log("QueryFilename completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	if err != nil {
		JsonAsResponse(w, err)
	} else {
		JsonAsResponse(w, response)
	}
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
