package routes

import (
	"encoding/json"
	logger "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"photo/album"
	"photo/database"
	"photo/exifhandler"
	"photo/folder"
	"photo/modele"
	"photo/slavehandler"
	"photo/webclient"
	"strconv"
	"time"
)

func CreateNewPhotoAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()

	albumCreationMessage := album.NewAlbumCreationMessage()

	err := json.NewDecoder(r.Body).Decode(albumCreationMessage)
	if err != nil {
		logger.Info("Cannot not decode body received for registering with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for registering", 400)
		return
	}
	logger.Info(albumCreationMessage)
	JsonAsResponse(w, "Album "+albumCreationMessage.AlbumName+" recorded.")
}

// route : return the extension files list from the configuration file
func GetExtensionList(w http.ResponseWriter, r *http.Request) {
	conf := modele.LoadConfigurationAtOnce()
	logger.Info("Ask for extension files")
	JsonAsResponse(w, conf)
}

//  route : purpose get the images files extension supported by the application
func ReadExtensionList(w http.ResponseWriter, r *http.Request) {
	logger.Info("get image files extension list")
	conf := slavehandler.GetSlaves()
	client := webclient.NewPhotoExifClient()
	if len(conf.Slaves) == 0 {
		JsonAsResponse(w, "Not clients registered")

	} else {
		var slave *slavehandler.Slave
		for _, slave = range conf.Slaves {
			break
		}
		if err, extensions := client.GetFileExtensionValues(slave); err != nil {
			JsonAsResponse(w, err.Error())
		} else {
			JsonAsResponse(w, extensions)
		}

	}
}

// route : purpose clean the database redundance
func CleanDatabase(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDatabaseHandler()
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	if err := db.CleanDatabase(); err != nil {
		JsonAsResponse(w, err)
		return
	}
	JsonAsResponse(w, "ok")
}

// route thumbnail  of the filpath (encoded in url)
func GetThumbnail(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("filepath")
	response, err := exifhandler.GetThumbnail(filePath)
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	JsonAsResponse(w, response)
}

// route: return the registered slaves on the controller
func GetRegisteredSlaves(w http.ResponseWriter, r *http.Request) {
	conf := slavehandler.GetSlaves()
	message := make([]modele.RegisteredSlave, 0)
	for _, slave := range conf.Slaves {
		message = append(message, modele.RegisteredSlave{MachineId: slave.Name, Ip: slave.Url})
	}
	logger.Infof("Ask for registered slave machines %v", message)
	JsonAsResponse(w, message)
}

// route: register a new slave that call this web service
func RegisterSlave(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()

	slave := &slavehandler.Slave{}
	err := json.NewDecoder(r.Body).Decode(slave)
	if err != nil {
		logger.Info("Cannot not decode body received for registering with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for registering", 400)
		return
	}
	slavehandler.AddSlave(slave)
	logger.Infof("%v %s\n", *slave, "is registered")
	JsonAsResponse(w, "ok")
}

// route : browse the directory (encoded in the value variable) on the machineId
func Browse(w http.ResponseWriter, r *http.Request) {
	//usr, _ := user.Current()
	starttime := time.Now()
	directorypath := r.URL.Query().Get("value")
	machineId := r.URL.Query().Get("machineId")
	logger.Info("Browse directory " + directorypath + " machineid " + machineId)
	response := &modele.DirectoryItemResponse{
		Name:             "Root",
		Path:             "#",
		MachineId:        machineId,
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
		logger.Error(err.Error())
	}
	logger.Info("Scan directory completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	JsonAsResponse(w, response)
}

func ScanFolders(w http.ResponseWriter, r *http.Request) {
	conf := modele.GetConfiguration()
	client := webclient.NewPhotoExifClient()
	folders := &modele.FolderToScan{}
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}

	err := json.NewDecoder(r.Body).Decode(folders)
	logger.Debug(&folders)
	var response string
	if err != nil {
		response = err.Error()
	} else {
		response = "Scans launched."
	}
	go client.ScanFoldersClient(folders.Folders, folders.MachineId, conf)

	JsonAsResponse(w, response)
}

func GetFileInformations(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	filepathValue := r.URL.Query().Get("value")
	logger.Info("file to scan " + filepathValue)
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.TagsPhoto, 0),
	}
	pinfos, err := exifhandler.GetPhotoInformations(filepathValue)
	if err != nil {
		response.Message = err.Error()
	}
	response.Photos = append(response.Photos, pinfos)
	logger.Info("Scan completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	logger.Info(strconv.Itoa(len(response.Photos)) + " images found")
	JsonAsResponse(w, response)
}

func GetDirectoryInformations(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.TagsPhoto, 0),
	}
	directorypath := r.URL.Query().Get("value")
	logger.Info("directory to scan " + directorypath)
	pinfos, err := exifhandler.GetPhotosInformations(directorypath, modele.LoadConfigurationAtOnce())
	response.Photos = pinfos
	if err != nil {
		response.Message = err.Error()
	}

	logger.Info("Scan completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	logger.Info(strconv.Itoa(len(response.Photos)) + " images found")
	JsonAsResponse(w, response)
}

func QueryExtension(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	filename := r.URL.Query().Get("value")
	size := r.URL.Query().Get("filesize")
	if size == "" {
		size = modele.FILESIZE_LITTLE
	}
	db, err := database.NewDatabaseHandler()
	if err != nil {
		logger.Error("Error while getting dabatabse with error" + err.Error())
		JsonAsResponse(w, err)
		return
	}
	response, err := db.QueryExtension(filename)
	response = database.Reduce(response, size)
	logger.Info("QueryExtension completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
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
	size := r.URL.Query().Get("filesize")
	if size == "" {
		size = modele.FILESIZE_LITTLE
	}
	db, err := database.NewDatabaseHandler()
	if err != nil {
		logger.Error("Error while getting dabatabse with error" + err.Error())
		JsonAsResponse(w, err)
		return
	}
	response, err := db.QueryExifTag(pattern, exiftag)
	response = database.Reduce(response, size)
	logger.Info("QueryExif completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	if err != nil {
		JsonAsResponse(w, err)
	} else {
		JsonAsResponse(w, response)
	}
}

func QueryFilename(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	filename := r.URL.Query().Get("value")
	size := r.URL.Query().Get("filesize")
	if size == "" {
		size = modele.FILESIZE_LITTLE
	}
	db, err := database.NewDatabaseHandler()
	if err != nil {
		logger.Error("Error while getting dabatabse with error" + err.Error())
		JsonAsResponse(w, err)
		return
	}
	response, err := db.QueryFilename(filename)
	response = database.Reduce(response, size)
	logger.Info("QueryFilename completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	if err != nil {
		JsonAsResponse(w, err)
	} else {
		JsonAsResponse(w, response)
	}
}

func QueryAll(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	db, err := database.NewDatabaseHandler()
	if err != nil {
		logger.Error("Error while getting dabatabse with error" + err.Error())
		JsonAsResponse(w, err)
		return
	}
	response, err := db.QueryAll()

	logger.Info("QueryAll completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	if err != nil {
		JsonAsResponse(w, err)
	} else {
		JsonAsResponse(w, response)
	}
}

func JsonAsResponse(w http.ResponseWriter, o interface{}) {
	js, err := json.Marshal(o)
	if err != nil {
		logger.Error("Error while marshalling  object")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application-json")
	w.Write(js)
}
