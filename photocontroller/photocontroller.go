package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"photo/modele"
	"photo/routes"
	"strconv"
	"time"
)

var httpport = flag.String("httpport", "", "listening at http://localhost:httpport")
var configurationfile = flag.String("configurationfile", "", "photoexif client's configuration file")
var Version string
var GitHash string
var BuildStmp string

func main() {

	flag.Parse()

	if *httpport != "" && *configurationfile != "" {
		modele.LoadPhotoExifConfiguration(*configurationfile)
		http.HandleFunc("/register", routes.RegisterSlave)
		http.HandleFunc("/registeredslaves", routes.GetRegisteredSlaves)
		http.HandleFunc("/browse", routes.Browse)
		http.HandleFunc("/scan", routes.ScanFolders)
		http.HandleFunc("/queryextension", routes.QueryExtension)
		http.HandleFunc("/queryfilename", routes.QueryFilename)
		http.HandleFunc("/queryexif", routes.QueryExif)
		http.HandleFunc("/queryall", routes.QueryAll)
		http.HandleFunc("/getfileextension", routes.ReadExtensionList)
		http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./resources"))))
		log.Fatal(http.ListenAndServe(":"+*httpport, nil))
	} else {
		timeStmp, err := strconv.Atoi(BuildStmp)
		if err != nil {
			timeStmp = 0
		}
		appVersion := "Version " + Version + ", build on " + time.Unix(int64(timeStmp), 0).String() + ", git hash " + GitHash
		fmt.Println(appVersion)
		flag.PrintDefaults()
	}

}
