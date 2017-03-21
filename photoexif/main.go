package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"photo/exifhandler"
	"photo/logger"
	"photo/modele"
	"photo/routes"
	"strconv"
	"time"
)

var photopath = flag.String("photopath", "", "photo path to analyze")
var directorypath = flag.String("directorypath", "", "directory path to scan.")
var httpport = flag.String("httpport", "", "listening at http://localhost:httpport")

func main() {
	conf := modele.LoadConfigurationAtOnce()
	starttime := time.Now()
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.TagsPhoto, 0),
	}
	flag.Parse()
	if *photopath != "" {
		response.Photos = append(response.Photos, exifhandler.GetPhotoInformations(*photopath))
	} else {
		if *directorypath != "" {
			response.Photos = exifhandler.GetPhotosInformations(*directorypath, conf)
		} else {
			if *httpport != "" {
				http.HandleFunc("/file", routes.GetFileInformations)
				http.HandleFunc("/directory", routes.GetDirectoryInformations)
				log.Fatal(http.ListenAndServe(":"+*httpport, nil))
			} else {
				flag.PrintDefaults()
			}
		}
	}
	logger.Log("Scan completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	logger.Log(strconv.Itoa(len(response.Photos)) + " images found")
	json.NewEncoder(os.Stdout).Encode(response)

}
