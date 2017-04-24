package database

import (
	"encoding/json"
	"github.com/HouzuoGuo/tiedot/db"
	logger "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"path/filepath"
	"photo/modele"
	"photo/slavehandler"
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
	DBPHOTO_COLLECTION = "photos_collection"
	DBALBUM_COLLECTION = "albums_collection"
	MACHINEID_INDEX    = "MachineId"
	FILENAME_INDEX     = "Filename"
	FILENAMES_INDEX    = "Filenames"
	FILEPATHS_INDEX    = "Filepaths"
	FILEPATH_INDEX     = "Filepath"
	MD5SUM_INDEX       = "Md5sum"
	FILETYPE_INDEX     = "Type"
	THUMBNAIL_INDEX    = "Thumbnail"
	EXIFTAGS_INDEX     = ""
)

func (d *DatabaseHandler) openDB() (*db.DB, error) {
	var err error
	var dbInstance *db.DB

	collectionExists := false
	databasePath := modele.GetConfiguration().DatabasePath
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
			break
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

	if err := feeds.Index([]string{MACHINEID_INDEX}); err != nil {
		logger.Error("Error while indexing MachineId with error : " + err.Error())
	}
	if err := feeds.Index([]string{FILENAME_INDEX}); err != nil {
		logger.Error("Error while indexing Filename with error : " + err.Error())
	}
	if err := feeds.Index([]string{FILENAMES_INDEX}); err != nil {
		logger.Error("Error while indexing Filenames with error : " + err.Error())
	}
	if err := feeds.Index([]string{FILEPATHS_INDEX}); err != nil {
		logger.Error("Error while indexing Filepaths with error : " + err.Error())
	}
	if err := feeds.Index([]string{FILEPATH_INDEX}); err != nil {
		logger.Error("Error while indexing Filepath with error : " + err.Error())
	}
	if err := feeds.Index([]string{MD5SUM_INDEX}); err != nil {
		logger.Error("Error while indexing Md5sum with error : " + err.Error())
	}
	if err := feeds.Index([]string{FILETYPE_INDEX}); err != nil {
		logger.Error("Error while indexing Type with error : " + err.Error())
	}
	if err := feeds.Index([]string{FILENAME_INDEX, FILEPATH_INDEX, FILETYPE_INDEX}); err != nil {
		logger.Error("Error while indexing Filename,Filepath,Type with error : " + err.Error())
	}

	return nil
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

func (d *DatabaseHandler) QueryAll() ([]*DatabasePhotoResponse, error) {
	response := make([]*DatabasePhotoResponse, 0)
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

func (d *DatabaseHandler) QueryExtension(pattern string) ([]*DatabasePhotoResponse, error) {
	response := make([]*DatabasePhotoResponse, 0)
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

func (d *DatabaseHandler) QueryFilename(pattern string) ([]*DatabasePhotoResponse, error) {
	response := make([]*DatabasePhotoResponse, 0)

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

		for _, val := range a[FILENAMES_INDEX].([]interface{}) {
			if strings.Contains(strings.ToLower(val.(string)), strings.ToLower(pattern)) {
				//logger.LogLn("Document",id,"is",string(docContent))
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
		return true
	})

	logger.Infof("request returns %d results for filename %s\n", len(response), pattern)
	return response, nil
}

func (d *DatabaseHandler) QueryExifTag(pattern string, exiftag string) ([]*DatabasePhotoResponse, error) {

	response := make([]*DatabasePhotoResponse, 0)
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

func Reduce(responses []*DatabasePhotoResponse, size string) []*DatabasePhotoResponse {

	finalResponses := make([]*DatabasePhotoResponse, 0)
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
