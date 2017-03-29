package main

import (
	"flag"
	"log"
	"net/http"
	"photo/modele"
	"photo/routes"
)

var httpport = flag.String("httpport", "", "listening at http://localhost:httpport")
var configurationfile = flag.String("configurationfile", "", "photoexif client's configuration file")

func main() {

	flag.Parse()

	if *httpport != "" && *configurationfile != "" {
		modele.LoadPhotoExifConfiguration(*configurationfile)
		http.HandleFunc("/browse", routes.Browse)
		http.HandleFunc("/scan", routes.ScanFolders)
		http.HandleFunc("/queryextension", routes.QueryExtension)
		http.HandleFunc("/queryfilename", routes.QueryFilename)
		http.HandleFunc("/queryexif", routes.QueryExif)
		http.HandleFunc("/queryall", routes.QueryAll)
		http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./resources"))))
		log.Fatal(http.ListenAndServe(":"+*httpport, nil))
	} else {
		flag.PrintDefaults()
	}

}
