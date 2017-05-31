package routes

import (
	"encoding/json"
	logger "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"photo/album"
	"photo/configurationapp"
	"photo/configurationexif"
	"photo/database"
	"photo/exifhandler"
	"photo/flickr_client"
	"photo/folder"
	"photo/google-photos_client"
	"photo/modele"
	"photo/pdf"
	"photo/slavehandler"
	"photo/webclient"
	"strconv"
	"strings"
	"time"
)

func GetHistory(w http.ResponseWriter, r *http.Request) {
	JsonAsResponse(w, modele.ActionsHistory)
}

func LoadFlickrConfiguration(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling load flickr configuration.")
	flickrconf := &flickr_client.Flickr{}
	flickrconf.LoadConfiguration()
	modele.PostActionMessage("calling load flickr configuration ended.")
	JsonAsResponse(w, flickrconf)
}

func SaveFlickrConfiguration(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling save flickr configuration.")
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()
	flickrconf := &flickr_client.Flickr{}
	err := json.NewDecoder(r.Body).Decode(flickrconf)
	if err != nil {
		logger.Info("Cannot not decode body received for flickr client with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for flickr client", 400)
		return
	}

	flickrClient := flickr_client.GetCurrentFlickrClient()
	flickrClient.ApiKey = flickrconf.ApiKey
	flickrClient.ApiSecret = flickrconf.ApiSecret
	flickrClient.Connect()
	flickrClient.GetUrlRequestToken()
	flickrconf.UrlAuthorization = flickrClient.UrlAuthorization
	flickr_client.SaveCurrentFlickrClient(flickrClient)
	modele.PostActionMessage("calling save flickr configuration ended.")
	JsonAsResponse(w, flickrconf)
}

func LoadFlickrAlbums(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling load flickr albums.")
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()
	flickrconf := &flickr_client.Flickr{}
	err := json.NewDecoder(r.Body).Decode(flickrconf)
	if err != nil {
		logger.Info("Cannot not decode body received for flickr client with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for flickr client", 400)
		return
	}

	// import data from google account
	go func() {
		flickrClient := flickr_client.GetCurrentFlickrClient()
		logger.Info(flickrconf)
		logger.Info(flickrClient)
		flickrClient.FlickrToken = flickrconf.FlickrToken
		data := flickrClient.GetData()
		db, err := database.NewDatabaseHandler()
		if err != nil {
			logger.Errorf("cannot connect to database with error %v", err)
			return
		}
		for _, response := range data {
			if err := db.InsertNewData(response); err != nil {
				logger.Errorf("cannot import google data into database with error %v", err)
			}
			md5sums := make([]string, 0)
			for _, photo := range response.Photos {
				md5sums = append(md5sums, photo.Md5Sum)
			}
			msg := album.NewAlbumMessage()
			msg.AlbumName = response.Origin
			msg.Md5sums = md5sums
			if err := db.InsertNewAlbum(msg); err != nil {
				logger.Errorf("cannot import google data into database with error %v", err)
			}
		}
		logger.Info("Import flickr albums finished")
		modele.PostActionMessage("calling load flickr albums ended.")
	}()
	JsonAsResponse(w, "Configuration saved and imported data from your flickr account")
}

func LoadGoogleConfiguration(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling load google configuration.")
	conf := configurationapp.GetConfiguration()
	googleconf := &google_photos_client.GooglePhotoClient{
		UserID: conf.GoogleUser,
		ID:     conf.GoogleID,
		Secret: conf.GoogleSecret,
	}
	logger.Info(googleconf)
	modele.PostActionMessage("calling load google configuration ended.")
	JsonAsResponse(w, googleconf)
}

func SaveGoogleConfiguration(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling save google configuration and load albums from google photos.")
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()
	googleConf := &google_photos_client.GooglePhotoClient{}
	err := json.NewDecoder(r.Body).Decode(googleConf)
	if err != nil {
		logger.Info("Cannot not decode body received for google client with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for google client", 400)
		return
	}
	conf := configurationapp.GetConfiguration()
	conf.GoogleID = googleConf.ID
	conf.GoogleSecret = googleConf.Secret
	conf.GoogleUser = googleConf.UserID

	if err := conf.Save(); err != nil {
		http.Error(w, "cannot save google client configuration", 400)
		return
	}

	logger.Info(googleConf)

	// import data from google account
	go func() {
		if err := googleConf.Connect(); err != nil {
			logger.Errorf("cannot connect to google photo account with error %v", err)
			return
		}
		data := googleConf.GetData()
		db, err := database.NewDatabaseHandler()
		if err != nil {
			logger.Errorf("cannot connect to database with error %v", err)
			return
		}
		for _, response := range data {
			if err := db.InsertNewData(response); err != nil {
				logger.Errorf("cannot import google data into database with error %v", err)
			}
			md5sums := make([]string, 0)
			for _, photo := range response.Photos {
				md5sums = append(md5sums, photo.Md5Sum)
			}
			msg := album.NewAlbumMessage()
			msg.AlbumName = response.Origin
			msg.Md5sums = md5sums
			if err := db.InsertNewAlbum(msg); err != nil {
				logger.Errorf("cannot import google data into database with error %v", err)
			}
		}
		modele.PostActionMessage("calling save google configuration and load albums from google photos ended.")
	}()

	JsonAsResponse(w, "Configuration saved and imported data, please check log file to accept account usage")
}

// route create a new album by the name and the md5sums of the photos
func CreateNewPhotoAlbum(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling create new album.")
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()

	albumMessage := album.NewAlbumMessage()

	err := json.NewDecoder(r.Body).Decode(albumMessage)
	if err != nil {
		logger.Info("Cannot not decode body received for registering with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for registering", 400)
		return
	}

	logger.Info(albumMessage)
	db, err := database.NewDatabaseHandler()
	modele.PostActionMessage("calling create new album ended.")
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	if err = db.InsertNewAlbum(albumMessage); err != nil {
		JsonAsResponse(w, err)
		return
	}

	JsonAsResponse(w, "Album "+albumMessage.AlbumName+" recorded.")
}
func DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling delete album content.")
	defer modele.PostActionMessage("calling delete album content ended.")
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()

	albumMessage := album.NewAlbumMessage()

	err := json.NewDecoder(r.Body).Decode(albumMessage)
	if err != nil {
		logger.Info("Cannot not decode body received for registering with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for registering", 400)
		return
	}
	logger.Info(albumMessage)

	db, err := database.NewDatabaseHandler()
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	if err = db.DeleteAlbum(albumMessage); err != nil {

		JsonAsResponse(w, err)
		return
	}

	JsonAsResponse(w, "Album "+albumMessage.AlbumName+" deleted.")

}

func DeletePhotosAlbum(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling delete photos from album content.")
	defer modele.PostActionMessage("calling delete photos from album content ended.")
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()

	albumMessage := album.NewAlbumMessage()

	err := json.NewDecoder(r.Body).Decode(albumMessage)
	if err != nil {
		logger.Info("Cannot not decode body received for registering with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for registering", 400)
		return
	}
	logger.Info(albumMessage)
	db, err := database.NewDatabaseHandler()
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	if err = db.DeletePhotoAlbum(albumMessage); err != nil {
		JsonAsResponse(w, err)
		return
	}

	JsonAsResponse(w, "Album "+albumMessage.AlbumName+" updated.")
}

func GenerateAlbumPdf(w http.ResponseWriter, r *http.Request) {

	albumName := r.URL.Query().Get("albumName")
	c, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		JsonAsResponse(w, "An error occured while generating pdf for album :"+albumName+" "+err.Error())
		return
	}

	photosid := strings.Split(c["photosid"][0], ",")
	modele.PostActionMessage("calling generate pdf album for album : " + albumName)

	logger.Info("Generate album : " + albumName)
	db, err := database.NewDatabaseHandler()
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	content := db.GetAlbumData(albumName)
	if content.AlbumName == albumName && len(content.Records) > 0 {
		selected := database.NewDatabaseAlbumRecord()
		logger.Infof("photosid:%v", photosid)
		if len(photosid) > 0 {
			selected.AlbumName = content.AlbumName
			for _, id := range photosid {
				for _, photo := range content.Records {
					if photo.Md5sum == id {
						selected.Records = append(selected.Records, photo)
						break
					}
				}
			}
		} else {
			selected = content
		}

		logger.Info(selected)
		photosFilenames := webclient.NewRawPhotoClient(selected).GetRemoteRawPhotosAlbum()
		data := pdf.CreatePdfAlbum(content.AlbumName, photosFilenames, pdf.Images3XPerPages)
		modele.PostActionMessage("calling generate pdf album for album : " + albumName + " ended.")
		BinaryAsResponse(w, data, albumName+".pdf")
		return
	}
	JsonAsResponse(w, "An error occured while generating pdf for album :"+albumName)
}

func UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling update album content.")
	if r.Body == nil {
		http.Error(w, "empty body", 400)
		return
	}
	defer r.Body.Close()

	albumMessage := album.NewAlbumMessage()

	err := json.NewDecoder(r.Body).Decode(albumMessage)
	if err != nil {
		logger.Info("Cannot not decode body received for registering with error " + err.Error())
		body, _ := ioutil.ReadAll(r.Body)
		logger.Debug("Body received : " + string(body))
		http.Error(w, "Cannot not decode body received for registering", 400)
		return
	}
	logger.Info(albumMessage)
	db, err := database.NewDatabaseHandler()
	modele.PostActionMessage("calling update album content ended.")
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	if err = db.UpdateAlbum(albumMessage); err != nil {
		JsonAsResponse(w, err)
		return
	}

	JsonAsResponse(w, "Album "+albumMessage.AlbumName+" updated.")

}

// return all albums names
func ListPhotoAlbums(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling get albums list.")
	db, err := database.NewDatabaseHandler()
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	albums := db.GetAlbumList()
	modele.PostActionMessage("calling get albums list ended.")
	JsonAsResponse(w, albums)
}

func GetAlbumData(w http.ResponseWriter, r *http.Request) {

	albumName := r.URL.Query().Get("albumName")
	modele.PostActionMessage("calling get album content for album : " + albumName)
	db, err := database.NewDatabaseHandler()
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	content := db.GetAlbumData(albumName)
	modele.PostActionMessage("calling get album content for album : " + albumName + " ended.")
	JsonAsResponse(w, content)
}

// route : return the extension files list from the configuration file
func GetExtensionList(w http.ResponseWriter, r *http.Request) {
	conf := configurationexif.LoadConfigurationAtOnce()
	logger.Info("Ask for extension files")
	JsonAsResponse(w, conf)
}

//  route : purpose get the images files extension supported by the application
func ReadExtensionList(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling get extension files")
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
	modele.PostActionMessage("calling get extension files ended.")
}

// route : purpose clean the database redundance
func CleanDatabase(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDatabaseHandler()
	modele.PostActionMessage("calling clean database.")
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	if err := db.CleanDatabase(); err != nil {
		JsonAsResponse(w, err)
		return
	}
	modele.PostActionMessage("calling clean database ended.")
	JsonAsResponse(w, "ok")
}

// route thumbnail  of the filpath (encoded in url)
func GetPhoto(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("filepath")
	content, orientation, err := exifhandler.GetBase64Photo(filePath)
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	JsonAsResponse(w, &modele.RawPhoto{Name: filePath, Data: content, Orientation: orientation})
}

// route thumbnail  of the filpath (encoded in url)
func GetThumbnail(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("filepath")
	response, err := exifhandler.GetBase64Thumbnail(filePath)
	if err != nil {
		JsonAsResponse(w, err)
		return
	}
	JsonAsResponse(w, response)
}

// route: return the registered slaves on the controller
func GetRegisteredSlaves(w http.ResponseWriter, r *http.Request) {
	modele.PostActionMessage("calling registered slaves.")
	conf := slavehandler.GetSlaves()
	message := make([]modele.RegisteredSlave, 0)
	for _, slave := range conf.Slaves {
		message = append(message, modele.RegisteredSlave{MachineId: slave.Name, Ip: slave.Url})
	}
	logger.Infof("Ask for registered slave machines %v", message)
	modele.PostActionMessage("exiting registered slaves.")
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
	modele.PostActionMessage("calling browse directory with directory : " + directorypath + " and machineid : " + machineId)
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
	modele.PostActionMessage("scan directory completed for directory : " + directorypath + " and machineid : " + machineId)
	JsonAsResponse(w, response)
}

func ScanFolders(w http.ResponseWriter, r *http.Request) {

	conf := configurationapp.GetConfiguration()
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
	modele.PostActionMessage("calling scan folders for machineid " + folders.MachineId)
	go client.ScanFoldersClient(folders.Folders, folders.MachineId, conf)

	JsonAsResponse(w, response)
}

func GetFileInformations(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	filepathValue := r.URL.Query().Get("value")
	logger.Info("file to scan " + filepathValue)
	response := &modele.PhotoResponse{
		Version: modele.VERSION,
		Photos:  make([]*modele.PhotoInformations, 0),
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
		Photos:  make([]*modele.PhotoInformations, 0),
	}
	directorypath := r.URL.Query().Get("value")
	logger.Info("directory to scan " + directorypath)
	pinfos, err := exifhandler.GetPhotosInformations(directorypath, configurationexif.LoadConfigurationAtOnce())
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
	modele.PostActionMessage("calling query extension with value : " + filename + " and filesize : " + size)
	db, err := database.NewDatabaseHandler()
	if err != nil {
		logger.Error("Error while getting dabatabse with error" + err.Error())
		JsonAsResponse(w, err)
		return
	}
	response, err := db.QueryExtension(filename)
	response = database.Reduce(response, size)
	logger.Info("QueryExtension completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	modele.PostActionMessage("calling query extension with value : " + filename + " and filesize : " + size + " ended.")
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
	modele.PostActionMessage("calling query exif with value : " + pattern + " and exiftag : " + exiftag + " and filesize : " + size)
	db, err := database.NewDatabaseHandler()
	if err != nil {
		logger.Error("Error while getting dabatabse with error" + err.Error())
		JsonAsResponse(w, err)
		return
	}
	response, err := db.QueryExifTag(pattern, exiftag)
	response = database.Reduce(response, size)
	logger.Info("QueryExif completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	modele.PostActionMessage("calling query exif with value : " + pattern + " and exiftag : " + exiftag + " and filesize : " + size + " ended.")
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
	modele.PostActionMessage("calling query filename with value : " + filename + " and filesize : " + size)
	db, err := database.NewDatabaseHandler()
	if err != nil {
		logger.Error("Error while getting dabatabse with error" + err.Error())
		JsonAsResponse(w, err)
		return
	}
	response, err := db.QueryFilename(filename)
	if err != nil {
		JsonAsResponse(w, err)
	}
	logger.Infof("QueryFilename returns %d records", len(response))
	response = database.Reduce(response, size)
	logger.Info("QueryFilename completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	modele.PostActionMessage("calling query filename with value : " + filename + " and filesize : " + size + " ended.")
	JsonAsResponse(w, response)
}

func QueryAll(w http.ResponseWriter, r *http.Request) {
	starttime := time.Now()
	modele.PostActionMessage("calling query all.")
	db, err := database.NewDatabaseHandler()
	if err != nil {
		logger.Error("Error while getting dabatabse with error" + err.Error())
		JsonAsResponse(w, err)
		return
	}
	response, err := db.QueryAll()

	logger.Info("QueryAll completed in " + strconv.FormatFloat(time.Now().Sub(starttime).Seconds(), 'g', 2, 64) + " seconds")
	modele.PostActionMessage("calling query ended.")
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

func BinaryAsResponse(w http.ResponseWriter, o []byte, filename string) {
	w.Header().Set("Content-type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write(o)
}
