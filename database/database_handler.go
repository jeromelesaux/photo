package database

import (
	"encoding/json"
	"github.com/HouzuoGuo/tiedot/db"
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
var DBCOLLECTION = "photos_collection"

func openDB() (*db.DB, error) {
	var err error
	createDB.Do(func() {
		collectionExists := false
		database, err = db.OpenDB("database_photo.db")
		if err != nil {
			logger.Log("Error while creating database with error : " + err.Error())
			return
		}
		for _, colname := range database.AllCols() {
			if colname == DBCOLLECTION {
				collectionExists = true
				break
			}
		}
		if !collectionExists {
			if err = database.Create(DBCOLLECTION); err != nil {
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
	var db *db.DB
	executeDBIndexes.Do(func() {
		db, err = openDB()
		if err != nil {
			logger.Log("Cannot use database with error : " + err.Error())
			return
		}
		feeds := db.Use(DBCOLLECTION)

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
	if strings.Contains(patternupper,"-") {
		result = append(result,strings.Split(patternupper,"-")...)
	}
	if strings.Contains(patternupper,"_") {
		result = append(result,strings.Split(patternupper,"_")...)
	}
	if strings.Contains(patternupper," ") {
		result = append(result,strings.Split(patternupper," ")...)
	}
	if strings.Contains(patternupper,".") {
		result = append(result,strings.Split(patternupper,".")...)
	}
	if strings.Contains(patternupper,string(filepath.Separator)) {
		result = append(result,strings.Split(patternupper,string(filepath.Separator))...)
	}

	return result
}

func InsertNewData(response *modele.PhotoResponse) error {
	db, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return err
	}

	feeds := db.Use(DBCOLLECTION)
	for _, item := range response.Photos {
		id, err := feeds.Insert(map[string]interface{}{
			"Filename": item.Filename,
			"Filenames":SplitAll(item.Filename),
			"Filepath" : item.Filepath,
			"Filepaths":SplitAll(item.Filepath),
			"Md5sum":   item.Md5Sum,
			"ExifTags": item.Tags,
			"Type":     strings.ToLower(filepath.Ext(item.Filename))})
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
		return nil, err
	}

	feeds := dataquery.Use(DBCOLLECTION)
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
				response[readBack["Md5sum"].(string)] = make(map[string]interface{},0)
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
		return nil, err
	}

	feeds := dataquery.Use(DBCOLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}

	json.Unmarshal([]byte(`[{"eq": "`+pattern+`", "in": ["Type"]}]`), &query)
	//json.Unmarshal([]byte(`["all"]`), &query)
	logger.LogLn(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Log("Error while querying with error :" + err.Error())
	}
	logger.Logf("request returns %d results\n",len(queryResult))
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Log("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			//logger.LogLn(readBack)
			response[readBack["Md5sum"].(string)] = readBack["ExifTags"]
			if readBack["ExifTags"] == nil {
				response[readBack["Md5sum"].(string)] = make(map[string]interface{},0)
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
		return nil, err
	}

	feeds := dataquery.Use(DBCOLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}

	json.Unmarshal([]byte(`[{"eq": "`+pattern+`", "in": ["Filenames"]}]`), &query)
	//json.Unmarshal([]byte(`["all"]`), &query)
	logger.LogLn(query)
	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Log("Error while querying with error :" + err.Error())
	}
	logger.Logf("request returns %d results\n",len(queryResult))
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Log("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			//logger.LogLn(readBack)
			response[readBack["Md5sum"].(string)] = readBack["ExifTags"]
			if readBack["ExifTags"] == nil {
				response[readBack["Md5sum"].(string)] = make(map[string]interface{},0)
			}
			response[readBack["Md5sum"].(string)].(map[string]interface{})["Filename"] = readBack["Filename"]
			response[readBack["Md5sum"].(string)].(map[string]interface{})["Filepath"] = readBack["Filepath"]
		}

	}

	return response, nil
}