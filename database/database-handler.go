package database

import (
	"encoding/json"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/pkg/errors"
	"path/filepath"
	"photo/logger"
	"photo/modele"
	"strconv"
	"strings"
	"sync"
)

var _ DatabaseInterface = (*DatabaseHandler)(nil)

type DatabaseHandler struct {
}

func NewDatabaseHandler() (*DatabaseHandler, error) {

	var err error
	databaseTiedotHandler := &DatabaseHandler{}
	createDB.Do(func() {
		if _, err = databaseTiedotHandler.openDB(); err != nil {
			return
		}
		if err = databaseTiedotHandler.createIndexes(); err != nil {
			return
		}
	})

	return databaseTiedotHandler, err
}

var createDB sync.Once

var DBPHOTOCOLLECTION = "photos_collection"
var DBALBUMCOLLECTION = "albums_collection"

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
		logger.Log("Error while creating database with error : " + err.Error())
		return dbInstance, err
	}

	for _, colname := range dbInstance.AllCols() {
		if colname == DBPHOTOCOLLECTION {
			collectionExists = true
			break
		}
	}
	if !collectionExists {
		if err = dbInstance.Create(DBPHOTOCOLLECTION); err != nil {
			logger.Log("Error while creating collection photos_collection with error : " + err.Error())
			return dbInstance, err
		} else {
			logger.Log("Creating collection " + DBPHOTOCOLLECTION)
		}
	}

	return dbInstance, err
}

func (d *DatabaseHandler) createIndexes() error {
	var err error
	var dbInstance *db.DB

	dbInstance, err = d.openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return err
	}

	feeds := dbInstance.Use(DBPHOTOCOLLECTION)

	if err := feeds.Index([]string{"MachineId"}); err != nil {
		logger.Log("Error while indexing MachineId with error : " + err.Error())
	}
	if err := feeds.Index([]string{"Filename"}); err != nil {
		logger.Log("Error while indexing Filename with error : " + err.Error())
	}
	if err := feeds.Index([]string{"Filenames"}); err != nil {
		logger.Log("Error while indexing Filenames with error : " + err.Error())
	}
	if err := feeds.Index([]string{"Filepaths"}); err != nil {
		logger.Log("Error while indexing Filepaths with error : " + err.Error())
	}
	if err := feeds.Index([]string{"Filepath"}); err != nil {
		logger.Log("Error while indexing Filepath with error : " + err.Error())
	}
	if err := feeds.Index([]string{"Md5sum"}); err != nil {
		logger.Log("Error while indexing Md5sum with error : " + err.Error())
	}
	if err := feeds.Index([]string{"Type"}); err != nil {
		logger.Log("Error while indexing Type with error : " + err.Error())
	}
	if err := feeds.Index([]string{"Filename", "Filepath", "Type"}); err != nil {
		logger.Log("Error while indexing Filename,Filepath,Type with error : " + err.Error())
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
		logger.Log("Error while opening database during insert operation with error " + err.Error())
		return err
	}

	feeds := dbInstance.Use(DBPHOTOCOLLECTION)
	for _, item := range response.Photos {
		id, err := feeds.Insert(map[string]interface{}{
			"MachineId": response.MachineId,
			"Filename":  item.Filename,
			"Filenames": SplitAll(item.Filename),
			"Filepath":  item.Filepath,
			"Filepaths": SplitAll(item.Filepath),
			"Md5sum":    item.Md5Sum,
			"ExifTags":  item.Tags,
			"Type":      strings.ToLower(filepath.Ext(item.Filename))})
		if err != nil {
			logger.Log("Cannot insert data in database with error : " + err.Error())
		} else {
			logger.Logf("DB return id %d for filepath:%s\n", id, item.Filepath)
		}

	}

	return nil
}

func (d *DatabaseHandler) QueryAll() ([]*DatabasePhotoResponse, error) {
	response := make([]*DatabasePhotoResponse, 0)
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Log("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}

	feeds := dbInstance.Use(DBPHOTOCOLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`["all"]`), &query)
	logger.LogLn(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Log("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Log("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			logger.LogLn(readBack)
			var exif map[string]interface{}
			if readBack["ExifTags"] != nil {
				exif = readBack["ExifTags"].(map[string]interface{})
			}
			response = append(response, NewDatabasePhotoResponse(
				readBack["Md5sum"].(string),
				readBack["Filename"].(string),
				readBack["Filepath"].(string),
				exif))
		}

	}

	return response, nil
}

func (d *DatabaseHandler) QueryExtenstion(pattern string) ([]*DatabasePhotoResponse, error) {
	response := make([]*DatabasePhotoResponse, 0)
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Log("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}

	feeds := dbInstance.Use(DBPHOTOCOLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}

	json.Unmarshal([]byte(`[{"eq": "`+pattern+`", "in": ["Type"]}]`), &query)
	//json.Unmarshal([]byte(`["all"]`), &query)
	logger.LogLn(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Log("Error while querying with error :" + err.Error())
	}
	logger.Logf("request returns %d results for extenstion %s\n", len(queryResult), pattern)
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Log("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			//logger.LogLn(readBack)
			var exif map[string]interface{}
			if readBack["ExifTags"] != nil {
				exif = readBack["ExifTags"].(map[string]interface{})
			}
			response = append(response, NewDatabasePhotoResponse(
				readBack["Md5sum"].(string),
				readBack["Filename"].(string),
				readBack["Filepath"].(string),
				exif))
		}

	}

	return response, nil
}

func (d *DatabaseHandler) QueryFilename(pattern string) ([]*DatabasePhotoResponse, error) {
	response := make([]*DatabasePhotoResponse, 0)

	dbInstance, err := d.openDB()
	if err != nil {
		logger.Log("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}

	feeds := dbInstance.Use(DBPHOTOCOLLECTION)

	feeds.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		var a map[string]interface{}
		err := json.Unmarshal(docContent, &a)
		if err != nil {
			logger.Log("Error while unmarshalling document with error : " + err.Error())
			return false
		}

		for _, val := range a["Filenames"].([]interface{}) {
			if strings.Contains(strings.ToLower(val.(string)), strings.ToLower(pattern)) {
				//logger.LogLn("Document",id,"is",string(docContent))
				var exif map[string]interface{}
				if a["ExifTags"] != nil {
					exif = a["ExifTags"].(map[string]interface{})
				}
				response = append(response, NewDatabasePhotoResponse(
					a["Md5sum"].(string),
					a["Filename"].(string),
					a["Filepath"].(string),
					exif))
			}

		}
		for _, val := range a["Filepaths"].([]interface{}) {
			if strings.Contains(strings.ToLower(val.(string)), strings.ToLower(pattern)) {
				//logger.LogLn("Document",id,"is",string(docContent))
				var exif map[string]interface{}
				if a["ExifTags"] != nil {
					exif = a["ExifTags"].(map[string]interface{})
				}
				response = append(response, NewDatabasePhotoResponse(
					a["Md5sum"].(string),
					a["Filename"].(string),
					a["Filepath"].(string),
					exif))
			}

		}

		//logger.LogLn("Document",id,"is",string(docContent))
		return true
		return false
	})

	logger.Logf("request returns %d results for filename %s\n", len(response), pattern)
	return response, nil
}

func (d *DatabaseHandler) QueryExifTag(pattern string, exiftag string) ([]*DatabasePhotoResponse, error) {

	response := make([]*DatabasePhotoResponse, 0)
	dbInstance, err := d.openDB()
	if err != nil {
		logger.Log("Error while opening database during insert operation with error " + err.Error())
		return response, err
	}

	feeds := dbInstance.Use(DBPHOTOCOLLECTION)
	feeds.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		var a map[string]interface{}
		err := json.Unmarshal(docContent, &a)
		if err != nil {
			logger.Log("Error while unmarshalling document with error : " + err.Error())
			return false
		}
		if a["ExifTags"] != nil {
			for key, val := range a["ExifTags"].(map[string]interface{}) {
				if strings.Contains(strings.ToLower(key), strings.ToLower(exiftag)) {
					if strings.Contains(strings.ToLower(val.(string)), strings.ToLower(pattern)) {
						var exif map[string]interface{}
						if a["ExifTags"] != nil {
							exif = a["ExifTags"].(map[string]interface{})
						}
						response = append(response, NewDatabasePhotoResponse(
							a["Md5sum"].(string),
							a["Filename"].(string),
							a["Filepath"].(string),
							exif))
					}
				}
			}
		}
		return true
		return false
	})

	logger.Logf("request returns %d results for pattern %s and exif tag %s\n", len(response), pattern, exiftag)
	return response, nil
}
