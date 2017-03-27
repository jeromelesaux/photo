package main

import (
	"flag"
	"log"
	"net/http"
	"photo/routes"
)

var httpport = flag.String("httpport", "", "listening at http://localhost:httpport")

func main() {

	flag.Parse()

	if *httpport != "" {
		http.HandleFunc("/browse", routes.Browse)
		http.HandleFunc("/scan", routes.ScanFolders)
		http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./resources"))))
		log.Fatal(http.ListenAndServe(":"+*httpport, nil))
	} else {
		flag.PrintDefaults()
	}

}
