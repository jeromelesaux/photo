package main

import (
	"flag"
	"fmt"
	logger "github.com/Sirupsen/logrus"
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
var wellcomeMessage = "\n" +
	"\\ \\        / / | | |                          | |\n" +
	" \\ \\  /\\  / /__| | | ___ ___  _ __ ___   ___  | |_ ___\n" +
	"  \\ \\/  \\/ / _ \\ | |/ __/ _ \\| '_ ` _ \\ / _ \\ | __/ _ \\ \n" +
	"   \\  /\\  /  __/ | | (_| (_) | | | | | |  __/ | || (_) |\n" +
	"    \\/  \\/ \\___|_|_|\\___\\___/|_| |_| |_|\\___|  \\__\\___/\n" +
	"\n" +
	"\n" +
	"__     ___    _ _\n" +
	"\\ \\   / / |  | | |        /\\ \n" +
	" \\ \\_/ /| |  | | |       /  \\ \n" +
	"  \\   / | |  | | |      / /\\ \\ \n" +
	"   | |  | |__| | |____ / ____ \\ \n" +
	"   |_|   \\____/|______/_/    \\_\\ \n\n"

func main() {

	flag.Parse()

	if *httpport != "" && *configurationfile != "" {
		logger.Info(wellcomeMessage)
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
		http.HandleFunc("/cleandatabase", routes.CleanDatabase)
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
