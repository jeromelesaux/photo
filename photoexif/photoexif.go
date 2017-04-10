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
	"photo/slavehandler"
	"strconv"
	"time"
)

var photopath = flag.String("photopath", "", "photo path to analyze")
var directorypath = flag.String("directorypath", "", "directory path to scan.")
var httpport = flag.String("httpport", "", "listening at http://localhost:httpport")
var masteruri = flag.String("masteruri", "", "uri of the master to register ex: -masteruri http://localhost:3001/register")

func main() {
	conf := modele.LoadConfigurationAtOnce()
	starttime := time.Now()
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.TagsPhoto, 0),
	}
	flag.Parse()
	if *photopath != "" {
		pinfos, err := exifhandler.GetPhotoInformations(*photopath)
		if err != nil {
			logger.Logf("Error with message :%s", err.Error())
		}
		response.Photos = append(response.Photos, pinfos)
	} else {
		if *directorypath != "" {
			pinfos, err := exifhandler.GetPhotosInformations(*directorypath, conf)
			if err != nil {
				logger.Logf("Error with message :%s", err.Error())
			}
			response.Photos = pinfos
		} else {
			if *httpport != "" {
				if *masteruri != "" {
					port, err := strconv.Atoi(*httpport)
					if err != nil {
						logger.Log("Error : " + err.Error())
						return
					}
					go slavehandler.RegisterToMaster(*masteruri, port, "/directory")
				} else {
					logger.Log("masteruri is mandatary, don't start")
					return
				}
				http.HandleFunc("/file", routes.GetFileInformations)
				http.HandleFunc("/directory", routes.GetDirectoryInformations)
				http.HandleFunc("/getfileextension", routes.GetExtensionList)
				http.HandleFunc("/thumbnail", routes.GetThumbnail)
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
