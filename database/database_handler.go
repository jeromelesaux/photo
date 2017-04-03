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

var createDB sync.Once
var executeDBIndexes sync.Once
var database *db.DB
var DBPHOTOCOLLECTION = "photos_collection"
var DBALBUMCOLLECTION = "albums_collection"

func openDB() (*db.DB, error) {
	var err error
	createDB.Do(func() {
		collectionExists := false
		databasePath := modele.GetConfiguration().DatabasePath
		if databasePath == "" {
			err = errors.New("No database path defined")
			return
		}
		database, err = db.OpenDB(databasePath)
		if err != nil {
			logger.Log("Error while creating database with error : " + err.Error())
			return
		}
		for _, colname := range database.AllCols() {
			if colname == DBPHOTOCOLLECTION {
				collectionExists = true
				break
			}
		}
		if !collectionExists {
			if err = database.Create(DBPHOTOCOLLECTION); err != nil {
				logger.Log("Error while creating collection photos_collection with error : " + err.Error())
				return
			}
		}
		return
	})
	return database, err
}

func createIndexes() error {
	var err error
	var databaseCI *db.DB
	executeDBIndexes.Do(func() {
		databaseCI, err = openDB()
		if err != nil {
			logger.Log("Cannot use database with error : " + err.Error())
			return
		}
		feeds := databaseCI.Use(DBPHOTOCOLLECTION)

		if err := feeds.Index([]string{"MachineId"}); err != nil {
			logger.Log("Error while indexing MachineId with error : " + err.Error())
			return
		}
		if err := feeds.Index([]string{"Filename"}); err != nil {
			logger.Log("Error while indexing Filename with error : " + err.Error())
			return
		}
		if err := feeds.Index([]string{"Filenames"}); err != nil {
			logger.Log("Error while indexing Filenames with error : " + err.Error())
			return
		}
		if err := feeds.Index([]string{"Filepaths"}); err != nil {
			logger.Log("Error while indexing Filepaths with error : " + err.Error())
			return
		}
		if err := feeds.Index([]string{"Filepath"}); err != nil {
			logger.Log("Error while indexing Filepath with error : " + err.Error())
			return
		}
		if err := feeds.Index([]string{"Md5sum"}); err != nil {
			logger.Log("Error while indexing Md5sum with error : " + err.Error())
			return
		}
		if err := feeds.Index([]string{"Type"}); err != nil {
			logger.Log("Error while indexing Type with error : " + err.Error())
			return
		}
		if err := feeds.Index([]string{"Filename", "Filepath", "Type"}); err != nil {
			logger.Log("Error while indexing Filename,Filepath,Type with error : " + err.Error())
			return
		}
	})
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

func InsertNewData(response *modele.PhotoResponse) error {
	databaseForInsert, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return err
	}

	feeds := databaseForInsert.Use(DBPHOTOCOLLECTION)
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

	if err := createIndexes(); err != nil {
		return err
	}

	return nil
}

func QueryAll() (map[string]interface{}, error) {
	response := make(map[string]interface{}, 0)
	dataquery, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return response, err
	}

	feeds := dataquery.Use(DBPHOTOCOLLECTION)
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
			response[readBack["Md5sum"].(string)] = readBack["ExifTags"]
			if readBack["ExifTags"] == nil {
				response[readBack["Md5sum"].(string)] = make(map[string]interface{}, 0)
			}
			response[readBack["Md5sum"].(string)].(map[string]interface{})["Filename"] = readBack["Filename"]
			response[readBack["Md5sum"].(string)].(map[string]interface{})["Filepath"] = readBack["Filepath"]
			response[readBack["Md5sum"].(string)].(map[string]interface{})["Filenames"] = readBack["Filenames"]
			response[readBack["Md5sum"].(string)].(map[string]interface{})["Filepaths"] = readBack["Filepaths"]
		}

	}
	return response, nil
}

func QueryExtenstion(pattern string) (map[string]interface{}, error) {
	response := make(map[string]interface{}, 0)
	dataquery, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return response, err
	}

	feeds := dataquery.Use(DBPHOTOCOLLECTION)
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
			response[readBack["Md5sum"].(string)] = readBack["ExifTags"]
			if readBack["ExifTags"] == nil {
				response[readBack["Md5sum"].(string)] = make(map[string]interface{}, 0)
			}
			response[readBack["Md5sum"].(string)].(map[string]interface{})["Filename"] = readBack["Filename"]
			response[readBack["Md5sum"].(string)].(map[string]interface{})["Filepath"] = readBack["Filepath"]

		}

	}

	return response, nil
}

func QueryFilename(pattern string) (map[string]interface{}, error) {
	response := make(map[string]interface{}, 0)
	dataquery, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return response, err
	}

	feeds := dataquery.Use(DBPHOTOCOLLECTION)

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
				response[a["Md5sum"].(string)] = a["ExifTags"]
				if response[a["Md5sum"].(string)] == nil {
					response[a["Md5sum"].(string)] = make(map[string]interface{}, 0)
				}
				response[a["Md5sum"].(string)].(map[string]interface{})["Filename"] = a["Filename"]
				response[a["Md5sum"].(string)].(map[string]interface{})["Filepath"] = a["Filepath"]

			}

		}
		for _, val := range a["Filepaths"].([]interface{}) {
			if strings.Contains(strings.ToLower(val.(string)), strings.ToLower(pattern)) {
				//logger.LogLn("Document",id,"is",string(docContent))
				response[a["Md5sum"].(string)] = a["ExifTags"]
				if response[a["Md5sum"].(string)] == nil {
					response[a["Md5sum"].(string)] = make(map[string]interface{}, 0)
				}
				response[a["Md5sum"].(string)].(map[string]interface{})["Filename"] = a["Filename"]
				response[a["Md5sum"].(string)].(map[string]interface{})["Filepath"] = a["Filepath"]

			}

		}

		//logger.LogLn("Document",id,"is",string(docContent))
		return true
		return false
	})
	logger.Logf("request returns %d results for filename %s\n", len(response), pattern)
	return response, nil
}

func QueryExifTag(pattern string, exiftag string) (map[string]interface{}, error) {

	response := make(map[string]interface{}, 0)
	dataquery, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return response, err
	}

	feeds := dataquery.Use(DBPHOTOCOLLECTION)
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
						response[a["Md5sum"].(string)] = a["ExifTags"]
						if response[a["Md5sum"].(string)] == nil {
							response[a["Md5sum"].(string)] = make(map[string]interface{}, 0)
						}
						response[a["Md5sum"].(string)].(map[string]interface{})["Filename"] = a["Filename"]
						response[a["Md5sum"].(string)].(map[string]interface{})["Filepath"] = a["Filepath"]
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
