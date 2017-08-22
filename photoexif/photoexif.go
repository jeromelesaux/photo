package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/jeromelesaux/photo/configurationexif"
	"github.com/jeromelesaux/photo/exifhandler"
	"github.com/jeromelesaux/photo/logger"
	"github.com/jeromelesaux/photo/modele"
	"github.com/jeromelesaux/photo/routes"
	"github.com/jeromelesaux/photo/slavehandler"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"
)

var photopath = flag.String("photopath", "", "photo path to analyze")
var directorypath = flag.String("directorypath", "", "directory path to scan.")
var httpport = flag.String("httpport", "", "listening at http://localhost:httpport")
var masteruri = flag.String("masteruri", "", "uri of the master to register ex: -masteruri http://localhost:3001/register")
var logFormat = flag.String("logformat", "", "format of the log (text or logstash available).")
var logLevel = flag.String("loglevel", "", "level of the log (DEBUG, INFO, WARN ...).")
var Version string
var GitHash string
var BuildStmp string

func main() {
	conf := configurationexif.LoadConfigurationAtOnce()
	starttime := time.Now()
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.PhotoInformations, 0),
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		timeStmp, err := strconv.Atoi(BuildStmp)
		if err != nil {
			timeStmp = 0
		}
		appVersion := "Version " + Version + ", build on " + time.Unix(int64(timeStmp), 0).String() + ", git hash " + GitHash
		fmt.Println(appVersion)
		flag.PrintDefaults()
		return
	}
	if *logFormat != "" {
		if *logLevel != "" {
			if err := logger.InitLog(*logLevel, *logFormat); err != nil {
				fmt.Println("Error of the log initialisation: " + err.Error())
			}
		} else {
			if err := logger.InitLog(*logLevel, *logFormat); err != nil {
				fmt.Println("Error of th log initialisation: " + err.Error())
			}
		}
	} else {
		if err := logger.InitLog("DEBUG", logger.TextFormatter); err != nil {
			fmt.Println("Error of th log initialisation: " + err.Error())
		}
	}

	if *photopath != "" {
		pinfos, err := exifhandler.GetPhotoInformations(*photopath)
		if err != nil {
			logrus.Errorf("Error with message :%s", err.Error())
		}
		response.Photos = append(response.Photos, pinfos)
	} else {
		if *directorypath != "" {
			pinfos, err := exifhandler.GetPhotosInformations(*directorypath, conf)
			if err != nil {
				logrus.Errorf("Error with message :%s", err.Error())
			}
			response.Photos = pinfos
		} else {
			if *httpport != "" {
				if *masteruri != "" {
					port, err := strconv.Atoi(*httpport)
					if err != nil {
						logrus.Error("Error : " + err.Error())
						return
					}
					go slavehandler.RegisterToMaster(*masteruri, port, "/directory")
				} else {
					logrus.Error("masteruri is mandatary, don't start")
					return
				}
				http.HandleFunc("/file", routes.GetFileInformations)
				http.HandleFunc("/directory", routes.GetDirectoryInformations)
				http.HandleFunc("/getfileextension", routes.GetExtensionList)
				http.HandleFunc("/thumbnail", routes.GetThumbnail)
				http.HandleFunc("/photo", routes.GetPhoto)
				log.Fatal(http.ListenAndServe(":"+*httpport, nil))
			}
		}
	}
	logrus.Info("Scan completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	logrus.Info(strconv.Itoa(len(response.Photos)) + " images found")
	json.NewEncoder(os.Stdout).Encode(response)

}
