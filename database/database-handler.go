package database

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/jeromelesaux/photo/album"
	"github.com/jeromelesaux/photo/configurationapp"
	"github.com/jeromelesaux/photo/modele"
	"github.com/jeromelesaux/photo/slavehandler"
	"github.com/pkg/errors"
	logger "github.com/sirupsen/logrus"
)

var _ DatabaseInterface = (*DatabaseHandler)(nil)
var globalDBConnection *db.DB

type DatabaseHandler struct {
	DBConnection *db.DB
}

func (d *DatabaseHandler) Close() error {
	return d.DBConnection.Close()
}

func NewDatabaseHandler() (*DatabaseHandler, error) {

	var err error
	databaseTiedotHandler := &DatabaseHandler{}

	if err = databaseTiedotHandler.openDB(); err != nil {
		return databaseTiedotHandler, err
	}
	createDB.Do(func() {
		if err = databaseTiedotHandler.createIndexes(); err != nil {
			return
		}
	})

	return databaseTiedotHandler, err
}

var (
	createDB                    sync.Once
	AlbumAlreadyExists          = errors.New("Album already exists in database.")
	ErrorWhileRetreivingAlbum   = errors.New("Error while retrieving album in database.")
	PictureAlreadyExists        = errors.New("Picture already exists in database.")
	ErrorWhileRetreivingPicture = errors.New("Error while retrieving picture in database.")
)

const (
	DBPHOTO_COLLECTION      = "photos_collection"
	DBALBUM_COLLECTION      = "albums_collection"
	MACHINEID_INDEX         = "MachineId"
	FILENAME_INDEX          = "Filename"
	FILENAMES_INDEX         = "Filenames"
	FILEPATHS_INDEX         = "Filepaths"
	FILEPATH_INDEX          = "Filepath"
	MD5SUM_INDEX            = "Md5sum"
	FILETYPE_INDEX          = "Type"
	THUMBNAIL_INDEX         = "Thumbnail"
	ALBUM_INDEX             = "Album"
	ALBUM_ITEMS             = "Album_Items"
	ALBUM_DESCRIPTION       = "Album_Description"
	EXIFTAGS_INDEX          = ""
	LONGITUDEGOOGLETAG      = "longitude"
	LATITUDEGOOGLETAG       = "latitude"
	LONGITUDEFLICKRTAG      = "GPS Longitude"
	LATITUDEFLICKRTAG       = "GPS Latitude"
	LONGITUDEREFFLICKTAG    = "GPS Longitude Ref"
	LATITUDEREFFLICKTAG     = "GPS Latitude Ref"
	LONGITUDEEXIFTOOLTAG    = "Longitude"
	LATITUDEEXIFTOOLTAG     = "Latitude"
	LONGITUDEREFEXIFTOOLTAG = "East or West Longitude"
	LATITUDEREFEXIFTOOLTAG  = "North or South Latitude"
	DATEFLICKRTAG           = "Date and Time (Original)"
	DATEGOOGLETAG           = "timestamp"
	ALBUM_TAGS              = "Album tags"
)

func (d *DatabaseHandler) openDB() error {
	var err error

	collectionExists := false
	albumExists := false
	databasePath := configurationapp.GetConfiguration().DatabasePath
	if databasePath == "" {
		err = errors.New("No database path defined")
		return err
	}
	if globalDBConnection == nil {
		globalDBConnection, err = db.OpenDB(databasePath)
		if err != nil {
			logger.Error("Error while creating database with error : " + err.Error())
			return err
		}
	}
	d.DBConnection = globalDBConnection

	for _, colname := range d.DBConnection.AllCols() {
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
		if err = d.DBConnection.Create(DBPHOTO_COLLECTION); err != nil {
			logger.Error("Error while creating collection photos_collection with error : " + err.Error())
			return err
		} else {
			logger.Info("Creating collection " + DBPHOTO_COLLECTION)
		}
	}

	if !albumExists {
		if err = d.DBConnection.Create(DBALBUM_COLLECTION); err != nil {
			logger.Error("Error while creating album photos_album with error : " + err.Error())
			return err
		} else {
			logger.Info("Creating album " + DBALBUM_COLLECTION)
		}
	}

	return err
}

func (d *DatabaseHandler) createIndexes() error {
	var err error

	if err != nil {
		logger.Error("Cannot use database with error : " + err.Error())
		return err
	}

	feedsPhoto := d.DBConnection.Use(DBPHOTO_COLLECTION)

	if err = feedsPhoto.Index([]string{MACHINEID_INDEX}); err != nil {
		logger.Error("Error while indexing MachineId with error : " + err.Error())
	}
	if err = feedsPhoto.Index([]string{FILENAME_INDEX}); err != nil {
		logger.Error("Error while indexing Filename with error : " + err.Error())
	}
	if err = feedsPhoto.Index([]string{FILENAMES_INDEX}); err != nil {
		logger.Error("Error while indexing Filenames with error : " + err.Error())
	}
	if err = feedsPhoto.Index([]string{FILEPATHS_INDEX}); err != nil {
		logger.Error("Error while indexing Filepaths with error : " + err.Error())
	}
	if err = feedsPhoto.Index([]string{FILEPATH_INDEX}); err != nil {
		logger.Error("Error while indexing Filepath with error : " + err.Error())
	}
	if err = feedsPhoto.Index([]string{MD5SUM_INDEX}); err != nil {
		logger.Error("Error while indexing Md5sum with error : " + err.Error())
	}
	if err = feedsPhoto.Index([]string{FILETYPE_INDEX}); err != nil {
		logger.Error("Error while indexing Type with error : " + err.Error())
	}
	if err = feedsPhoto.Index([]string{FILENAME_INDEX, FILEPATH_INDEX, FILETYPE_INDEX}); err != nil {
		logger.Errorf("Error while indexing Filename,Filepath,Type with error : %v", err.Error())
	}

	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
	if err = feedsAlbum.Index([]string{ALBUM_INDEX}); err != nil {
		logger.Error("Error while indexing Album with error : " + err.Error())
	}
	if err = feedsAlbum.Index([]string{ALBUM_TAGS}); err != nil {
		logger.Errorf("Error while indexing Albums tags with error %v", err)
	}
	if err = feedsAlbum.Index([]string{ALBUM_INDEX, ALBUM_TAGS}); err != nil {
		logger.Errorf("Error while indexing Albums tags with error %v", err)
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

func ToUnixTime(exif map[string]interface{}, groupby string) string {
	if exif[DATEGOOGLETAG] != nil {
		v, err := strconv.ParseInt(exif[DATEGOOGLETAG].(string), 10, 64)
		if err != nil {
			logger.Errorf("Error while parsing date %s, with error %v", exif[DATEGOOGLETAG].(string), err)
			return ""
		}
		tm := time.Unix((v / 1000), 0)
		switch groupby {
		case "month":
			return time.Date(tm.Year(), tm.Month(), 0, 0, 0, 0, 0, tm.Location()).Format("2006-01-02")
		case "year":
			return time.Date(tm.Year(), 0, 0, 0, 0, 0, 0, tm.Location()).Format("2006-01-02")
		default:
			return ""
		}
	} else {
		if exif[DATEFLICKRTAG] != nil {
			tm, err := time.Parse("2006:01:02 15:04:05", exif[DATEFLICKRTAG].(string))
			if err != nil {
				logger.Errorf("Error while parsing date %s, with error %v", exif[DATEFLICKRTAG].(string), err)
				return ""
			}
			switch groupby {
			case "month":
				return time.Date(tm.Year(), tm.Month(), 0, 0, 0, 0, 0, tm.Location()).Format("2006-01-02")
			case "year":
				return time.Date(tm.Year(), 0, 0, 0, 0, 0, 0, tm.Location()).Format("2006-01-02")
			default:
				return ""
			}
		}
	}
	return ""
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
		} else {
			if exif[LONGITUDEEXIFTOOLTAG] != nil && exif[LATITUDEEXIFTOOLTAG] != nil {
				var d, m, s float64
				var latitude, longitude float64
				_, err := fmt.Sscanf(exif[LATITUDEEXIFTOOLTAG].(string), "%f, %f, %f", &d, &m, &s)
				if err != nil {
					logger.Errorf("Error while parsing flickr Latitude string %s %v ", exif[LATITUDEEXIFTOOLTAG].(string), err)
				} else {
					latitude = Round(d+(m/60)+(s/3600), .5, 2)
				}
				_, err = fmt.Sscanf(exif[LONGITUDEEXIFTOOLTAG].(string), "%f, %f, %f", &d, &m, &s)
				if err != nil {
					logger.Errorf("Error while parsing flickr Longitude string %s %v", exif[LONGITUDEEXIFTOOLTAG].(string), err)
				} else {
					longitude = Round(d+(m/60)+(s/3600), .5, 2)
				}
				if exif[LATITUDEREFEXIFTOOLTAG].(string) == "S" {
					latitude *= -1
				}
				if exif[LONGITUDEREFEXIFTOOLTAG].(string) == "W" {
					longitude *= -1
				}
				return Round(latitude, .5, 2), Round(longitude, .5, 2)
			}
		}
	}
	return .0, .0
}

func (d *DatabaseHandler) QueryByTag(tag string) ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)

	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feedsAlbum, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
		return response, err
	}

	tagToFound := strings.TrimSpace(strings.ToUpper(tag))

	for id := range queryResult {
		readBack, err := feedsAlbum.Read(id)
		if err != nil {
			logger.Errorf("Error while retrieving id %d with error : %v", id, err.Error())
		} else {
			tagExists := false
			logger.Infof("Album name : %v Tags :%v", readBack[ALBUM_INDEX], readBack[ALBUM_TAGS])
			if readBack[ALBUM_TAGS] != nil {
				for _, t := range readBack[ALBUM_TAGS].([]interface{}) {
					if strings.TrimSpace(strings.ToUpper(t.(string))) == tagToFound {
						tagExists = true
						break
					}
				}
			}
			if tagExists {
				albumName := readBack[ALBUM_INDEX].(string)
				albumRecord := d.GetAlbumData(albumName)
				for _, v := range albumRecord.Records {
					response = append(response, v)
				}
			}
		}
	}

	return response, nil
}

func (d *DatabaseHandler) GetOriginStats() (*album.OriginStatsMessage, error) {
	o := album.NewOriginStatsMessage()

	feedsCollection := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
			logger.Errorf("Error while retrieving id %d with error : %v", id, err.Error())
		} else {
			var machineOrigin = readBack[MACHINEID_INDEX].(string)
			o.Stats[machineOrigin]++
		}

	}
	return o, nil
}

func (d *DatabaseHandler) GetPhotosUrl(md5sums []string) ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)

	for _, m := range md5sums {
		queryResult := make(map[int]struct{})
		var query interface{}
		json.Unmarshal([]byte(`[{"eq": "`+m+`", "in": ["`+MD5SUM_INDEX+`"]}]`), &query)
		logger.Info(query)
		if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
			logger.Error("Error while querying with error :" + err.Error())
		}
		for id := range queryResult {
			readBack, err := feeds.Read(id)
			if err != nil {
				logger.Errorf("Error while retrieving id %d by md5sum with error : %v", id, err.Error())
			} else {
				logger.Debug(readBack)
			}
			origin := readBack[MACHINEID_INDEX].(string)
			filepath := readBack[FILEPATH_INDEX].(string)
			if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
				filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", readBack[FILEPATH_INDEX].(string), origin)
			}
			response = append(response, NewDatabasePhotoResponse(
				readBack[MD5SUM_INDEX].(string),
				readBack[FILENAME_INDEX].(string),
				filepath,
				readBack[MACHINEID_INDEX].(string),
				"",
				nil))
		}
	}
	return Reduce(response, ""), nil
}

func (d *DatabaseHandler) GetPhotosFromTime(queryDate string, groupby string) ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
			logger.Errorf("Error while retrieving id %d with error : %v", id, err.Error())
		} else {
			logger.Debug(readBack)
		}
		var exif map[string]interface{}
		if readBack[EXIFTAGS_INDEX] != nil {
			exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
			photoDate := ToUnixTime(exif, groupby)
			if photoDate == queryDate {
				origin := readBack[MACHINEID_INDEX].(string)
				filepath := readBack[FILEPATH_INDEX].(string)
				if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
					filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", readBack[FILEPATH_INDEX].(string), origin)
				}
				response = append(response, NewDatabasePhotoResponse(
					readBack[MD5SUM_INDEX].(string),
					readBack[FILENAME_INDEX].(string),
					filepath,
					readBack[MACHINEID_INDEX].(string),
					readBack[THUMBNAIL_INDEX].(string),
					exif))
			}
		}

	}
	return response, nil
}

func (d *DatabaseHandler) GetTimeStats(groupby string) (*album.TimeStatsMessage, error) {
	response := album.NewTimeStatsMessage()

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
			logger.Errorf("Error while retreiveing id %d with error : %v", id, err.Error())
		} else {
			logger.Debug(readBack)
		}
		var exif map[string]interface{}
		if readBack[EXIFTAGS_INDEX] != nil {
			exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
			timeStat := &album.TimeStatMessage{Date: ToUnixTime(exif, groupby), Count: 1}
			var found = false
			for _, v := range response.Stats {
				if v.Date == timeStat.Date {
					found = true
					v.Count++
					break
				}
			}
			if !found {
				response.Stats = append(response.Stats, timeStat)
			}
		}

	}

	return response, nil
}

func (d *DatabaseHandler) GetPhotosFromCoordinates(lat, lng string) ([]*DatabasePhotoRecord, error) {
	qlatitude, _ := strconv.ParseFloat(lat, 64)
	qlongitude, _ := strconv.ParseFloat(lng, 64)
	response := make([]*DatabasePhotoRecord, 0)

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
			logger.Errorf("Error while retreiveing id %d with error : %v", id, err.Error())
		} else {
			logger.Debug(readBack)
			var exif map[string]interface{}
			if readBack[EXIFTAGS_INDEX] != nil {
				exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
				latitude, longitude := CoordinatesFromExif(exif)
				if Round(longitude, .5, 2) == qlongitude && Round(latitude, .5, 2) == qlatitude {
					origin := readBack[MACHINEID_INDEX].(string)
					filepath := readBack[FILEPATH_INDEX].(string)
					if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
						filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", readBack[FILEPATH_INDEX].(string), origin)
					}
					response = append(response, NewDatabasePhotoResponse(
						readBack[MD5SUM_INDEX].(string),
						readBack[FILENAME_INDEX].(string),
						filepath,
						readBack[MACHINEID_INDEX].(string),
						readBack[THUMBNAIL_INDEX].(string),
						exif))
				}
			}
		}
	}

	return response, nil
}

func (d *DatabaseHandler) GetLocationStats() (*album.LocationStatsMessage, error) {
	l := album.NewLocationStatsMessage()

	feedsPhotos := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
			logger.Errorf("Error while retreiveing id %d with error : %v", id, err.Error())
		} else {
			var exif map[string]interface{}
			if readBack[EXIFTAGS_INDEX] != nil {
				exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
				latitude, longitude := CoordinatesFromExif(exif)
				if longitude != 0. && latitude != 0. {
					gps := &album.LocationMessage{
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
			}

		}
	}

	return l, nil

}

func (d *DatabaseHandler) GetAlbumList() []string {
	albumsNames := make([]string, 0)

	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
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
			logger.Errorf("Error while retreiveing id %d with error : %v", id, err.Error())
		} else {
			name := readBack[ALBUM_INDEX].(string)
			alreadyInSlice := false
			for _, v := range albumsNames {
				if v == name {
					alreadyInSlice = true
					break
				}
			}
			if !alreadyInSlice {
				albumsNames = append(albumsNames, name)
			}
		}
	}
	return albumsNames
}

func (d *DatabaseHandler) AlbumExists(albumName string) (bool, error) {

	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}

	json.Unmarshal([]byte(`[{"eq": "`+albumName+`", "in": ["`+ALBUM_INDEX+`"]}]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feedsAlbum, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	if len(queryResult) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (d *DatabaseHandler) GetAlbumData(albumName string) *DatabaseAlbumRecord {
	collection := NewDatabaseAlbumRecord()
	collection.AlbumName = albumName

	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
	feedsCollection := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
			logger.Errorf("Error while retreiveing id %d  with error : %v", id, err.Error())
		} else {
			logger.Infof("description : %s", readBack[ALBUM_DESCRIPTION].(string))
			collection.Description = readBack[ALBUM_DESCRIPTION].(string)
			if readBack[ALBUM_TAGS] != nil {

				tags := readBack[ALBUM_TAGS].([]interface{})
				for _, v := range tags {
					collection.Tags = append(collection.Tags, v.(string))
					logger.Infof("Tag :%s", v.(string))
				}
			}
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
						logger.Errorf("Error while retreiveing id %d with error : %v ", id, err.Error())
					} else {
						logger.Debug(readBack)
						var exif map[string]interface{}
						if readBack[EXIFTAGS_INDEX] != nil {
							exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
						}
						origin := readBack[MACHINEID_INDEX].(string)
						filepath := readBack[FILEPATH_INDEX].(string)
						if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
							filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", readBack[FILEPATH_INDEX].(string), origin)
						}
						collection.Records = append(collection.Records,
							&DatabasePhotoRecord{
								MachineId: readBack[MACHINEID_INDEX].(string),
								Md5sum:    readBack[MD5SUM_INDEX].(string),
								Filename:  readBack[FILENAME_INDEX].(string),
								Filepath:  filepath,
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

	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
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
			logger.Errorf("Error while retreiveing id  %d with error : %v", id, err.Error())
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
				ALBUM_TAGS:        response.Tags,
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

	exists, err := d.AlbumExists(response.AlbumName)
	if err != nil {
		return err
	}

	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)

	if exists {
		return d.UpdateAlbum(response)
	} else {

		id, err := feedsAlbum.Insert(map[string]interface{}{
			ALBUM_INDEX:       response.AlbumName,
			ALBUM_ITEMS:       response.Md5sums,
			ALBUM_DESCRIPTION: response.Description,
			ALBUM_TAGS:        response.Tags,
		})
		if err != nil {
			logger.Errorf("Cannot insert album %s in database with error : %v", response.AlbumName, err)
		} else {
			logger.Infof("DB return id %d for album:%s\n", id, response.AlbumName)
		}
	}

	return err
}

func (d *DatabaseHandler) DeleteAlbum(response *album.AlbumMessage) error {

	var err error
	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
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

	var err error
	md5sumsMerged := make([]string, 0)

	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
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
			logger.Errorf("Error while retreiveing id %d with error : %v", id, err.Error())
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
				ALBUM_TAGS:        response.Tags,
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

func (d *DatabaseHandler) PictureExists(md5sum string) (bool, error) {

	feedsCollection := d.DBConnection.Use(DBPHOTO_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`[{"eq": "`+md5sum+`", "in": ["`+MD5SUM_INDEX+`"]}]`), &query)
	if err := db.EvalQuery(query, feedsCollection, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
		return true, ErrorWhileRetreivingPicture
	}
	if len(queryResult) == 0 {
		return false, nil
	}

	return true, nil
}

func (d *DatabaseHandler) InsertNewData(response *modele.PhotoResponse) error {

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
	for _, item := range response.Photos {
		exists, err := d.PictureExists(item.Md5Sum)
		if !exists && err == nil {
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
		} else {
			if err == nil {
				logger.Infof("This picture %s already exists in database skipped.", item.Md5Sum)
			} else {
				logger.Infof("Error for %s with error %v", item.Md5Sum, err)
			}
		}

	}

	return nil
}

func (d *DatabaseHandler) removeDuplicateAlbums() error {

	// suppress album more than 1
	feedsAlbum := d.DBConnection.Use(DBALBUM_COLLECTION)
	var queryAlbum interface{}
	json.Unmarshal([]byte(`["all"]`), &queryAlbum)
	queryResultAlbum := make(map[int]struct{})
	logger.Info(queryAlbum)
	if err := db.EvalQuery(queryAlbum, feedsAlbum, &queryResultAlbum); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	// suppress pictures more than 1
	for id := range queryResultAlbum {
		readBack, err := feedsAlbum.Read(id)
		if err != nil {
			logger.Errorf("Error while retrieving id %d  with error : ", id, err.Error())
		}
		albumName := readBack[ALBUM_INDEX]
		if albumName != nil {
			var subquery interface{}
			json.Unmarshal([]byte(`[{"eq": "`+albumName.(string)+`", "in": ["`+ALBUM_INDEX+`"]}]`), &subquery)
			subqueryResult := make(map[int]struct{})
			if err := db.EvalQuery(subquery, feedsAlbum, &queryResultAlbum); err != nil {
				logger.Error("Error while querying with error :" + err.Error())
			}
			if len(subqueryResult) > 1 {
				index := 0
				for id := range subqueryResult {
					readBack, err := feedsAlbum.Read(id)
					if err != nil {
						logger.Errorf("Error while retrieving id %d  with error : %v ", id, err.Error())
					} else {
						if index == 0 {
							logger.Infof("Keeping album %d %s", id, readBack[ALBUM_INDEX])
						} else {
							logger.Infof("Deleting album %d %s", id, readBack[ALBUM_INDEX])
							feedsAlbum.Delete(id)
						}
					}
					index++
				}
			}
		}
	}
	return nil
}

func (d *DatabaseHandler) removeDuplicatePhotos() error {

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
	var query interface{}
	queryResult := make(map[int]struct{})
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	// suppress pictures more than 1
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Errorf("Error while retreiveing id %d  with error : ", id, err.Error())
		}
		md5sum := readBack[MD5SUM_INDEX]
		if md5sum != nil {
			var subquery interface{}
			json.Unmarshal([]byte(`[{"eq": "`+md5sum.(string)+`", "in": ["`+MD5SUM_INDEX+`"]}]`), &subquery)
			subqueryResult := make(map[int]struct{})
			if err := db.EvalQuery(subquery, feeds, &subqueryResult); err != nil {
				logger.Error("Error while querying with error :" + err.Error())
			}
			if len(subqueryResult) > 1 {
				index := 0
				for id := range subqueryResult {
					readBack, err := feeds.Read(id)
					if err != nil {
						logger.Errorf("Error while retreiveing id %d  with error : ", id, err.Error())
					} else {
						if index == 0 {
							logger.Infof("Keeping image %d %s", id, readBack[MD5SUM_INDEX])
						} else {
							logger.Infof("Deleting image %d %s", id, readBack[MD5SUM_INDEX])
							feeds.Delete(id)
						}
					}
					index++
				}
			}
		}
	}
	return nil
}

func (d *DatabaseHandler) CleanDatabase() error {
	slaves := slavehandler.GetSlaves()

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	queryResult := make(map[int]struct{})

	logger.Info(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Errorf("Error while retreiveing id %d with error : %v", id, err.Error())
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

	if err := d.removeDuplicatePhotos(); err != nil {
		logger.Errorf("Error while removing duplicates photos with error %v", err)
	}

	if err := d.removeDuplicateAlbums(); err != nil {
		logger.Errorf("Error while removing duplicates albums with error %v", err)
	}

	if err := d.DBConnection.Scrub(DBPHOTO_COLLECTION); err != nil {
		logger.Errorf("Error while scrubbing collection %s with error %v", DBPHOTO_COLLECTION, err)
		return err
	}

	if err := d.DBConnection.Scrub(DBALBUM_COLLECTION); err != nil {
		return err
	}

	return nil
}

func (d *DatabaseHandler) QueryAll() ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
			logger.Errorf("Error while retreiveing id %d with error : %v", id, err.Error())
		} else {
			logger.Debug(readBack)
			var exif map[string]interface{}
			if readBack[EXIFTAGS_INDEX] != nil {
				exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
			}
			origin := readBack[MACHINEID_INDEX].(string)
			filepath := readBack[FILEPATH_INDEX].(string)
			if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
				filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", readBack[FILEPATH_INDEX].(string), origin)
			}
			response = append(response, NewDatabasePhotoResponse(
				readBack[MD5SUM_INDEX].(string),
				readBack[FILENAME_INDEX].(string),
				filepath,
				readBack[MACHINEID_INDEX].(string),
				readBack[THUMBNAIL_INDEX].(string),
				exif))
		}

	}
	return response, nil
}

func (d *DatabaseHandler) QueryExtension(pattern string) ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)
	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}

	json.Unmarshal([]byte(`[{"eq": "`+pattern+`", "in": ["`+FILETYPE_INDEX+`"]}]`), &query)
	logger.Info(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Error("Error while querying with error :" + err.Error())
	}
	logger.Infof("request returns %d results for extension %s\n", len(queryResult), pattern)
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Errorf("Error while retrieving id %d with error : %v", id, err.Error())
		} else {
			//logger.LogLn(readBack)
			var exif map[string]interface{}
			if readBack[EXIFTAGS_INDEX] != nil {
				exif = readBack[EXIFTAGS_INDEX].(map[string]interface{})
			}
			origin := readBack[MACHINEID_INDEX].(string)
			filepath := readBack[FILEPATH_INDEX].(string)
			if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
				filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", readBack[FILEPATH_INDEX].(string), origin)
			}
			response = append(response, NewDatabasePhotoResponse(
				readBack[MD5SUM_INDEX].(string),
				readBack[FILENAME_INDEX].(string),
				filepath,
				readBack[MACHINEID_INDEX].(string),
				readBack[THUMBNAIL_INDEX].(string),
				exif))
		}

	}

	return response, nil
}

func (d *DatabaseHandler) QueryFilename(pattern string) ([]*DatabasePhotoRecord, error) {
	response := make([]*DatabasePhotoRecord, 0)

	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
					origin := a[MACHINEID_INDEX].(string)
					filepath := a[FILEPATH_INDEX].(string)
					if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
						filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", a[FILEPATH_INDEX].(string), origin)
					}
					response = append(response, NewDatabasePhotoResponse(
						a[MD5SUM_INDEX].(string),
						a[FILENAME_INDEX].(string),
						filepath,
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
					origin := a[MACHINEID_INDEX].(string)
					filepath := a[FILEPATH_INDEX].(string)
					if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
						filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", a[FILEPATH_INDEX].(string), origin)
					}
					response = append(response, NewDatabasePhotoResponse(
						a[MD5SUM_INDEX].(string),
						a[FILENAME_INDEX].(string),
						filepath,
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
	feeds := d.DBConnection.Use(DBPHOTO_COLLECTION)
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
						origin := a[MACHINEID_INDEX].(string)
						filepath := a[FILEPATH_INDEX].(string)
						if origin != modele.ORIGIN_FLICKR && origin != modele.ORIGIN_GOOGLE {
							filepath = fmt.Sprintf("/photo?filepath=%s&machineid=%s", a[FILEPATH_INDEX].(string), origin)
						}
						response = append(response, NewDatabasePhotoResponse(
							a[MD5SUM_INDEX].(string),
							a[FILENAME_INDEX].(string),
							filepath,
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
	finalResponses.Tags = album.Tags
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
