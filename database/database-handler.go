package database

import (
	"encoding/json"
	"fmt"
	"github.com/HouzuoGuo/tiedot/db"
	logger "github.com/Sirupsen/logrus"
	"github.com/jeromelesaux/photo/album"
	"github.com/jeromelesaux/photo/configurationapp"
	"github.com/jeromelesaux/photo/modele"
	"github.com/jeromelesaux/photo/slavehandler"
	"github.com/pkg/errors"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var _ DatabaseInterface = (*DatabaseHandler)(nil)

type DatabaseHandler struct {
}

func NewDatabaseHandler() (*DatabaseHandler, error) {
	var dbinstance *db.DB
	var err error
	databaseTiedotHandler := &DatabaseHandler{}
	createDB.Do(func() {
		if dbinstance, err = databaseTiedotHandler.openDB(); err != nil {
			return
		}
		dbinstance.Close()
		if err = databaseTiedotHandler.createIndexes(); err != nil {
			return
		}
	})

	return databaseTiedotHandler, err
}

var createDB sync.Once

var (
	DBPHOTO_COLLECTION   = "photos_collection"
	DBALBUM_COLLECTION   = "albums_collection"
	MACHINEID_INDEX      = "MachineId"
	FILENAME_INDEX       = "Filename"
	FILENAMES_INDEX      = "Filenames"
	FILEPATHS_INDEX      = "Filepaths"
	FILEPATH_INDEX       = "Filepath"
	MD5SUM_INDEX         = "Md5sum"
	FILETYPE_INDEX       = "Type"
	THUMBNAIL_INDEX      = "Thumbnail"
	ALBUM_INDEX          = "Album"
	ALBUM_ITEMS          = "Album_Items"
	ALBUM_DESCRIPTION    = "Album_Description"
	EXIFTAGS_INDEX       = ""
	LONGITUDEGOOGLETAG   = "longitude"
	LATITUDEGOOGLETAG    = "latitude"
	LONGITUDEFLICKRTAG   = "GPS Longitude"
	LATITUDEFLICKRTAG    = "GPS Latitude"
	LONGITUDEREFFLICKTAG = "GPS Longitude Ref"
	LATITUDEREFFLICKTAG  = "GPS Latitude Ref"
)

func (d *DatabaseHandler) openDB() (*db.DB, error) {
	var err error
	var dbInstance *db.DB

	collectionExists := false
	albumExists := false
	databasePath := configurationapp.GetConfiguration().DatabasePath
	if databasePath == "" {
		err = errors.New("No database path defined")
		return dbInstance, err
	}
	dbInstance, err = db.OpenDB(databasePath)
	if err != nil {
		logger.Error("Error while creating database with error : " + err.Error())
		return dbInstance, err
	}

	for _, colname := range dbInstance.AllCols() {
		if colname == DBPHOTO_COLLECTION {
			collectionExists = true
			if albumExists == true {
				break
			}
		}
		if colname == DBALBUM_COLLECTION {
			albumExists = true
			if collectionExists == true {
				break
			}
		}
	}
	if !collectionExists {
		if err = dbInstance.Create(DBPHOTO_COLLECTION); err != nil {
			logger.Error("Error while creating collection photos_collection with error : " + err.Error())
			return dbInstance, err
		} else {
			logger.Info("Creating collection " + DBPHOTO_COLLECTION)
		}
	}

	if !albumExists {
		if err = dbInstance.Create(DBALBUM_COLLECTION); err != nil {
			logger.Error("Error while creating album photos_album with error : " + err.Error())
			return dbInstance, err
		} else {
			logger.Info("Creating album " + DBALBUM_COLLECTION)
		}
	}

	return dbInstance, err
}

func (d *DatabaseHandler) createIndexes() error {
	var err error
	var dbInstance *db.DB

	dbInstance, err = d.openDB()
	if err != nil {
		logger.Error("Cannot use database with error : " + err.Error())
		return err
	}
	defer dbInstance.Close()

	feeds := dbInstance.Use(DBPHOTO_COLLECTION)

	if err = feeds.Index([]string{MACHINEID_INDEX}); err != nil {
		logger.Error("Error while indexing MachineId with error : " + err.Error())
	}
	if err = feeds.Index([]string{FILENAME_INDEX}); err != nil {
		logger.Error("Error while indexing Filename with error : " + err.Error())
	}
	if err = feeds.Index([]string{FILENAMES_INDEX}); err != nil {
		logger.Error("Error while indexing Filenames with error : " + err.Error())
	}
	if err = feeds.Index([]string{FILEPATHS_INDEX}); err != nil {
		logger.Error("Error while indexing Filepaths with error : " + err.Error())
	}
	if err = feeds.Index([]string{FILEPATH_INDEX}); err != nil {
		logger.Error("Error while indexing Filepath with error : " + err.Error())
	}
	if err = feeds.Index([]string{MD5SUM_INDEX}); err != nil {
		logger.Error("Error while indexing Md5sum with error : " + err.Error())
	}
	if err = feeds.Index([]string{FILETYPE_INDEX}); err != nil {
		logger.Error("Error while indexing Type with error : " + err.Error())
	}
	if err = feeds.Index([]string{FILENAME_INDEX, FILEPATH_INDEX, FILETYPE_INDEX}); err != nil {
		logger.Error("Error while indexing Filename,Filepath,Type with error : " + err.Error())
	}

	feeds = dbInstance.Use(DBALBUM_COLLECTION)
	if err = feeds.Index([]string{ALBUM_INDEX}); err != nil {
		logger.Error("Error while indexing Album with error : " + err.Error())
	}

	return nil
}

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func SplitAll(pattern string) []string {
	var result []string
	patternupper := strings.ToUpper(pattern)
	if strings.Contains(patternupper, "-") {
		result = append(result, strings.Split(patternupper, "-")...)
	}
	if strings.Contains(patternupper, "_") {
		result = append(result, strings.Split(patternupper, "_")...)
	}
	if strings.Contains(patternupper, " ") {
		result = append(result, strings.Split(patternupper, " ")...)
	}
	if strings.Contains(patternupper, ".") {
		result = append(result, strings.Split(patternupper, ".")...)
	}
	if strings.Contains(patternupper, string(filepath.Separator)) {
		result = append(result, strings.Split(patternupper, string(filepath.Separator))...)
	}

	return result
}

func CoordinatesFromExif(exif map[string]interface{}) (float64, float64) {
	if exif[LONGITUDEGOOGLETAG] != nil && exif[LATITUDEGOOGLETAG] != nil {
		longitude, _ := strconv.ParseFloat(exif[LONGITUDEGOOGLETAG].(string), 64)
		latitude, _ := strconv.ParseFloat(exif[LATITUDEGOOGLETAG].(string), 64)
		return Round(latitude, .5, 2), Round(longitude, .5, 2)
	} else {
		if exif[LONGITUDEFLICKRTAG] != nil && exif[LATITUDEFLICKRTAG] != nil {
			var d, m, s float64
			var latitude, longitude float64
			_, err := fmt.Sscanf(exif[LATITUDEFLICKRTAG].(string), "%f deg %f' %f", &d, &m, &s)
			if err != nil {
				logger.Errorf("Error while parsing flickr Latitude string %s %v ", exif[LATITUDEFLICKRTAG].(string), err)
			} else {
				latitude = Round(d+(m/60)+(s/3600), .5, 2)
			}
			_, err = fmt.Sscanf(exif[LONGITUDEFLICKRTAG].(string), "%f deg %f' %f", &d, &m, &s)
			if err != nil {
				logger.Errorf("Error while parsing flickr Longitude string %s %v", exif[LONGITUDEFLICKRTAG].(string), err)
			} else {
				longitude = Round(d+(m/60)+(s/3600), .5, 2)
			}
			if exif[LATITUDEREFFLICKTAG].(string) == "South" {
				latitude *= -1
			}
			if exif[LONGITUDEREFFLICKTAG].(string) == "West" {
				longitude *= -1
			}
			return Round(latitude, .5, 2), Round(longitude, .5, 2)
		}
	}
	return .0, .0
}

func (d *DatabaseHandler) GetOriginStats() (*album.OriginStatsMessage, error) {
	o := album.NewOriginStatsMessage()
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return o, err
	}
	feedsCollection := dbInstance.Use(DBPHOTO_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feedsCollection, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feedsCollection.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			var machineOrigin = readBack[MACHINEID_INDEX].(string)
			o.Stats[machineOrigin]++
		}

	}

	dbInstance.Close()
	return o, nil
}

func (d *DatabaseHandler) GetPhotosFromCoordinates(lat, lng string) ([]*DatabasePhotoRecord, error) {
	qlatitude, _ := strconv.ParseFloat(lat, 64)
	qlongitude, _ := strconv.ParseFloat(lng, 64)
	response := make([]*DatabasePhotoRecord, 0)
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}

	feeds := dbInstance.Use(DBPHOTO_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			logger.Debug(readBack)
			var exif map[string]interface{}
			if readBack[EXIFTAGS_INDEX] != nil {
				exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})

				if exif[LONGITUDEGOOGLETAG] != nil && exif[LATITUDEGOOGLETAG] != nil {
					latitude, longitude := CoordinatesFromExif(exif)
					if Round(longitude, .5, 2) == qlongitude && Round(latitude, .5, 2) == qlatitude {
						response = append(response, NewDatabasePhotoResponse(
							readBack[MD5SUM_INDEX].(string),
							readBack[FILENAME_INDEX].(string),
							readBack[FILEPATH_INDEX].(string),
							readBack[MACHINEID_INDEX].(string),
							readBack[THUMBNAIL_INDEX].(string),
							exif))
					}
				} else {
					if exif[LONGITUDEFLICKRTAG] != nil && exif[LATITUDEFLICKRTAG] != nil {
						latitude, longitude := CoordinatesFromExif(exif)
						if Round(longitude, .5, 2) == qlongitude && Round(latitude, .5, 2) == qlatitude {
							response = append(response, NewDatabasePhotoResponse(
								readBack[MD5SUM_INDEX].(string),
								readBack[FILENAME_INDEX].(string),
								readBack[FILEPATH_INDEX].(string),
								readBack[MACHINEID_INDEX].(string),
								readBack[THUMBNAIL_INDEX].(string),
								exif))
						}

					}
				}
			}

		}

	}

	dbInstance.Close()
	return response, nil
}

func (d *DatabaseHandler) GetLocationStats() (*album.LocationStatsMessage, error) {
	l := album.NewLocationStatsMessage()
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return l, err
	}
	defer dbInstance.Close()
	feedsPhotos := dbInstance.Use(DBPHOTO_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feedsPhotos, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feedsPhotos.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			var exif map[string]interface{}
			if readBack[EXIFTAGS_INDEX] != nil {
				exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
				if exif[LONGITUDEGOOGLETAG] != nil && exif[LATITUDEGOOGLETAG] != nil {
					latitude, longitude := CoordinatesFromExif(exif)
					if longitude != 0. && latitude != 0. {
						gps := album.LocationMessage{
							Longitude: Round(longitude, .5, 2),
							Latitude:  Round(latitude, .5, 2)}
						var found = false
						for i, v := range l.Stats {
							if v.Longitude == gps.Longitude && v.Latitude == gps.Latitude {
								l.Stats[i].Count++
								found = true
								break
							}
						}
						if !found {
							gps.Count = 1
							l.Stats = append(l.Stats, gps)
						}
					}
				} else {
					if exif[LONGITUDEFLICKRTAG] != nil && exif[LATITUDEFLICKRTAG] != nil {
						latitude, longitude := CoordinatesFromExif(exif)
						if longitude != 0. && latitude != 0. {
							gps := album.LocationMessage{Latitude: latitude, Longitude: longitude}
							var found = false
							for i, v := range l.Stats {
								if v.Longitude == gps.Longitude && v.Latitude == gps.Latitude {
									l.Stats[i].Count++
									found = true
									break
								}
							}
							if !found {
								gps.Count = 1
								l.Stats = append(l.Stats, gps)
							}

						}

					}
				}
			}

		}
	}

	return l, nil

}

func (d *DatabaseHandler) GetAlbumList() []string {
	albumsNames := make([]string, 0)
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return albumsNames
	}
	defer dbInstance.Close()
	feedsAlbum := dbInstance.Use(DBALBUM_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feedsAlbum, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feedsAlbum.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			albumsNames = append(albumsNames, readBack[ALBUM_INDEX].(string))
		}
	}
	return albumsNames
}

func (d *DatabaseHandler) GetAlbumData(albumName string) *DatabaseAlbumRecord {
	collection := NewDatabaseAlbumRecord()
	collection.AlbumName = albumName
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return collection
	}
	defer dbInstance.Close()
	feedsAlbum := dbInstance.Use(DBALBUM_COLLECTION)
	feedsCollection := dbInstance.Use(DBPHOTO_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}

	json.Unmarshal([]byte(`[{"eq": "`+albumName+`", "in": ["`+ALBUM_INDEX+`"]}]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feedsAlbum, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}

	for id := range queryResult {
		readBack, err := feedsAlbum.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			logger.Infof("description : %s", readBack[ALBUM_DESCRIPTION].(string))
			collection.Description = readBack[ALBUM_DESCRIPTION].(string)
			for _, md5sum := range readBack[ALBUM_ITEMS].([]interface{}) {
				queryResultImg := make(map[int]struct{})
				var queryImg interface{}
				json.Unmarshal([]byte(`[{"eq": "`+md5sum.(string)+`", "in": ["`+MD5SUM_INDEX+`"]}]`), &queryImg)
				logger.Info(query)
				if err := db.EvalQuery(queryImg, feedsCollection, &queryResultImg); err != nil {
					logger.Error("Error while querying with error :" + err.Error())
				}
				for id := range queryResultImg {
					readBack, err := feedsCollection.Read(id)
					if err != nil {
						logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
					} else {
						logger.Debug(readBack)
						var exif map[string]interface{}
						if readBack[EXIFTAGS_INDEX] != nil {
							exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
						}
						collection.Records = append(collection.Records,
							&DatabasePhotoRecord{
								MachineId: readBack[MACHINEID_INDEX].(string),
								Md5sum:    readBack[MD5SUM_INDEX].(string),
								Filename:  readBack[FILENAME_INDEX].(string),
								Filepath:  readBack[FILEPATH_INDEX].(string),
								Thumbnail: readBack[THUMBNAIL_INDEX].(string),
								ExifTags:  exif,
							})
					}

				}
			}
		}
	}

	return ReduceAlbumMessage(collection, "")
}

func (d *DatabaseHandler) DeletePhotoAlbum(response *album.AlbumMessage) error {
	var err error
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return err
	}
	defer dbInstance.Close()

	feedsAlbum := dbInstance.Use(DBALBUM_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}

	json.Unmarshal([]byte(`[{"eq": "`+response.AlbumName+`", "in": ["`+ALBUM_INDEX+`"]}]`), &query)
	logger.Info(query)
	if err = db.EvalQuery(query, feedsAlbum, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	photosToKeep := make([]string, 0)
	for id := range queryResult {
		var readback map[string]interface{}
		readback, err = feedsAlbum.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			for _, item := range readback[ALBUM_ITEMS].([]interface{}) {
				mustBeDeleted := false
				for _, md5sum := range response.Md5sums {
					if md5sum == item.(string) {
						mustBeDeleted = true
						break
					}
				}
				if !mustBeDeleted {
					photosToKeep = append(photosToKeep, item.(string))
				}
			}
			err = feedsAlbum.Update(id, map[string]interface{}{
				ALBUM_INDEX:       response.AlbumName,
				ALBUM_ITEMS:       photosToKeep,
				ALBUM_DESCRIPTION: response.Description,
			})
			if err != nil {
				logger.Error("Cannot insert data in database with error : " + err.Error())
			} else {
				logger.Infof("DB return id %d for album:%s\n", id, response.AlbumName)
			}

		}
	}
	return err
}

func (d *DatabaseHandler) InsertNewAlbum(response *album.AlbumMessage) error {
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return err
	}
	defer dbInstance.Close()

	feedsAlbum := dbInstance.Use(DBALBUM_COLLECTION)
	id, err := feedsAlbum.Insert(map[string]interface{}{
		ALBUM_INDEX:       response.AlbumName,
		ALBUM_ITEMS:       response.Md5sums,
		ALBUM_DESCRIPTION: response.Description,
	})
	if err != nil {
		logger.Error("Cannot insert data in database with error : " + err.Error())
	} else {
		logger.Infof("DB return id %d for album:%s\n", id, response.AlbumName)
	}

	return err
}

func (d *DatabaseHandler) DeleteAlbum(response *album.AlbumMessage) error {
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return err
	}
	defer dbInstance.Close()

	feedsAlbum := dbInstance.Use(DBALBUM_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`[{"eq": "`+response.AlbumName+`", "in": ["`+ALBUM_INDEX+`"]}]`), &query)
	logger.Info(query)
	if err = db.EvalQuery(query, feedsAlbum, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
		return err
	}
	if len(queryResult) == 0 {
		return errors.New("no records found")
	}
	for id := range queryResult {
		err = feedsAlbum.Delete(id)

		if err != nil {
			logger.Error("Cannot delete data in database with error : " + err.Error())
		} else {
			logger.Infof("DB return id %d for album:%s is delete\n", id, response.AlbumName)
		}

	}
	return err
}

func (d *DatabaseHandler) UpdateAlbum(response *album.AlbumMessage) error {
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return err
	}
	defer dbInstance.Close()

	md5sumsMerged := make([]string, 0)

	feedsAlbum := dbInstance.Use(DBALBUM_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`[{"eq": "`+response.AlbumName+`", "in": ["`+ALBUM_INDEX+`"]}]`), &query)
	logger.Info(query)
	if err = db.EvalQuery(query, feedsAlbum, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
		return err
	}
	if len(queryResult) == 0 {
		return errors.New("no records found")
	}
	md5sumsMerged = append(md5sumsMerged, response.Md5sums...)
	for id := range queryResult {
		var readback map[string]interface{}
		readback, err = feedsAlbum.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			for _, item := range readback[ALBUM_ITEMS].([]interface{}) {
				existing := false
				for _, md5sum := range md5sumsMerged {
					if md5sum == item.(string) {
						existing = true
						break
					}
				}
				if !existing {
					md5sumsMerged = append(md5sumsMerged, item.(string))
				}
			}
			err = feedsAlbum.Update(id, map[string]interface{}{
				ALBUM_INDEX:       response.AlbumName,
				ALBUM_ITEMS:       md5sumsMerged,
				ALBUM_DESCRIPTION: response.Description,
			})
			if err != nil {
				logger.Error("Cannot insert data in database with error : " + err.Error())
			} else {
				logger.Infof("DB return id %d for album:%s\n", id, response.AlbumName)
			}
		}
	}
	return err
}

func (d *DatabaseHandler) InsertNewData(response *modele.PhotoResponse) error {
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return err
	}
	defer dbInstance.Close()

	feeds := dbInstance.Use(DBPHOTO_COLLECTION)
	for _, item := range response.Photos {
		id, err := feeds.Insert(map[string]interface{}{
			MACHINEID_INDEX: response.MachineId,
			FILENAME_INDEX:  item.Filename,
			FILENAMES_INDEX: SplitAll(item.Filename),
			FILEPATH_INDEX:  item.Filepath,
			FILEPATHS_INDEX: SplitAll(item.Filepath),
			MD5SUM_INDEX:    item.Md5Sum,
			EXIFTAGS_INDEX:  item.Tags,
			THUMBNAIL_INDEX: item.Thumbnail,
			FILETYPE_INDEX:  strings.ToLower(filepath.Ext(item.Filename))})
		if err != nil {
			logger.Error("Cannot insert data in database with error : " + err.Error())
		} else {
			logger.Infof("DB return id %d for filepath:%s\n", id, item.Filepath)
		}

	}

	return nil
}

func (d *DatabaseHandler) CleanDatabase() error {
	slaves := slavehandler.GetSlaves()
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return err
	}
	defer dbInstance.Close()
	feeds := dbInstance.Use(DBPHOTO_COLLECTION)

	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			if readBack[MACHINEID_INDEX] == "" {
				logger.Infof("Removing %d", id)
				feeds.Delete(id)
			}
			for _, slave := range slaves.Slaves {
				if !slave.IsActive() && slave.Name == readBack[MACHINEID_INDEX] {
					logger.Infof("Removing %d", id)
					feeds.Delete(id)
				}
			}
		}
	}

	return nil
}

func (d *DatabaseHandler) QueryAll() ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}

	feeds := dbInstance.Use(DBPHOTO_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			logger.Debug(readBack)
			var exif map[string]interface{}
			if readBack[EXIFTAGS_INDEX] != nil {
				exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
			}
			response = append(response, NewDatabasePhotoResponse(
				readBack[MD5SUM_INDEX].(string),
				readBack[FILENAME_INDEX].(string),
				readBack[FILEPATH_INDEX].(string),
				readBack[MACHINEID_INDEX].(string),
				readBack[THUMBNAIL_INDEX].(string),
				exif))
		}

	}

	dbInstance.Close()
	return response, nil
}

func (d *DatabaseHandler) QueryExtension(pattern string) ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}
	defer dbInstance.Close()

	feeds := dbInstance.Use(DBPHOTO_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}

	json.Unmarshal([]byte(`[{"eq": "`+pattern+`", "in": ["`+FILETYPE_INDEX+`"]}]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	logger.Infof("request returns %d results for extenstion %s\n", len(queryResult), pattern)
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Error("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			//logger.LogLn(readBack)
			var exif map[string]interface{}
			if readBack[EXIFTAGS_INDEX] != nil {
				exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
			}
			response = append(response, NewDatabasePhotoResponse(
				readBack[MD5SUM_INDEX].(string),
				readBack[FILENAME_INDEX].(string),
				readBack[FILEPATH_INDEX].(string),
				readBack[MACHINEID_INDEX].(string),
				readBack[THUMBNAIL_INDEX].(string),
				exif))
		}

	}

	return response, nil
}

func (d *DatabaseHandler) QueryFilename(pattern string) ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)

	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}
	defer dbInstance.Close()

	feeds := dbInstance.Use(DBPHOTO_COLLECTION)
	feeds.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		var a map[string]interface{}
		err := json.Unmarshal(docContent, &a)
		if err != nil {
			logger.Error("Error while unmarshalling document with error : " + err.Error())
			return false
		}

		if a[FILENAMES_INDEX] != nil {
			for _, val := range a[FILENAMES_INDEX].([]interface{}) {
				if strings.Contains(strings.ToLower(val.(string)), strings.ToLower(pattern)) {
					var exif map[string]interface{}
					if a[EXIFTAGS_INDEX] != nil {
						exif = a[EXIFTAGS_INDEX].(map[string]interface{})
					}
					response = append(response, NewDatabasePhotoResponse(
						a[MD5SUM_INDEX].(string),
						a[FILENAME_INDEX].(string),
						a[FILEPATH_INDEX].(string),
						a[MACHINEID_INDEX].(string),
						a[THUMBNAIL_INDEX].(string),
						exif))
				}

			}
		}
		if a[FILEPATHS_INDEX] != nil {
			for _, val := range a[FILEPATHS_INDEX].([]interface{}) {
				if strings.Contains(strings.ToLower(val.(string)), strings.ToLower(pattern)) {
					var exif map[string]interface{}
					if a[EXIFTAGS_INDEX] != nil {
						exif = a[EXIFTAGS_INDEX].(map[string]interface{})
					}
					response = append(response, NewDatabasePhotoResponse(
						a[MD5SUM_INDEX].(string),
						a[FILENAME_INDEX].(string),
						a[FILEPATH_INDEX].(string),
						a[MACHINEID_INDEX].(string),
						a[THUMBNAIL_INDEX].(string),
						exif))
				}

			}
		}
		return true
	})
	logger.Infof("request returns %d results for filename %s\n", len(response), pattern)
	return response, nil
}

func (d *DatabaseHandler) QueryExifTag(pattern string, exiftag string) ([]*DatabasePhotoRecord, error) {

	response := make([]*DatabasePhotoRecord, 0)
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Error("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}
	defer dbInstance.Close()

	feeds := dbInstance.Use(DBPHOTO_COLLECTION)
	feeds.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		var a map[string]interface{}
		err := json.Unmarshal(docContent, &a)
		if err != nil {
			logger.Error("Error while unmarshalling document with error : " + err.Error())
			return false
		}
		if a[EXIFTAGS_INDEX] != nil {
			for key, val := range a[EXIFTAGS_INDEX].(map[string]interface{}) {
				if strings.Contains(strings.ToLower(key), strings.ToLower(exiftag)) {
					if strings.Contains(strings.ToLower(val.(string)), strings.ToLower(pattern)) {
						var exif map[string]interface{}
						if a[EXIFTAGS_INDEX] != nil {
							exif = a[EXIFTAGS_INDEX].(map[string]interface{})
						}
						response = append(response, NewDatabasePhotoResponse(
							a[MD5SUM_INDEX].(string),
							a[FILENAME_INDEX].(string),
							a[FILEPATH_INDEX].(string),
							a[MACHINEID_INDEX].(string),
							a[THUMBNAIL_INDEX].(string),
							exif))
					}
				}
			}
		}
		return true
	})

	logger.Infof("request returns %d results for pattern %s and exif tag %s\n", len(response), pattern, exiftag)
	return response, nil
}

func Reduce(responses []*DatabasePhotoRecord, size string) []*DatabasePhotoRecord {

	finalResponses := make([]*DatabasePhotoRecord, 0)
	for _, response := range responses {
		alreadyStored := false
		for _, r := range finalResponses {
			if r.Md5sum == response.Md5sum {
				alreadyStored = true
				break
			}
		}
		if !alreadyStored {
			//logger.Infof("filesize :%d", len(data))
			switch size {
			case modele.FILESIZE_LITTLE:
				finalResponses = append(finalResponses, response)
			case modele.FILESIZE_MEDIUM:
				if len(response.Image) > 15000 {
					finalResponses = append(finalResponses, response)
				}
			case modele.FILESIZE_BIG:
				if len(response.Image) > 25000 {
					finalResponses = append(finalResponses, response)
				}
			default:
				finalResponses = append(finalResponses, response)
			}
		}
	}
	return finalResponses
}

func ReduceAlbumMessage(album *DatabaseAlbumRecord, size string) *DatabaseAlbumRecord {
	finalResponses := NewDatabaseAlbumRecord()
	finalResponses.AlbumName = album.AlbumName
	finalResponses.Description = album.Description
	for _, response := range album.Records {
		alreadyStored := false
		for _, r := range finalResponses.Records {
			if r.Md5sum == response.Md5sum {
				alreadyStored = true
				break
			}
		}
		if !alreadyStored {
			//logger.Infof("filesize :%d", len(data))
			switch size {
			case modele.FILESIZE_LITTLE:
				finalResponses.Records = append(finalResponses.Records, response)
			case modele.FILESIZE_MEDIUM:
				if len(response.Image) > 15000 {
					finalResponses.Records = append(finalResponses.Records, response)
				}
			case modele.FILESIZE_BIG:
				if len(response.Image) > 25000 {
					finalResponses.Records = append(finalResponses.Records, response)
				}
			default:
				finalResponses.Records = append(finalResponses.Records, response)
			}
		}
	}
	return finalResponses
}
