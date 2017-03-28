package database

import (
	"sync"
	"github.com/HouzuoGuo/tiedot/db"
	"photo/logger"
	"photo/modele"
	"encoding/json"
	"strconv"
)

var createDB sync.Once
var database *db.DB
var DBCOLLECTION = "photos_collection"

func openDB() (*db.DB, error) {
	var err error
	createDB.Do(func() {

		database, err = db.OpenDB("database_photo.db")
		if err != nil {
			logger.Log("Error while creating database with error : " + err.Error())
			return
		}
		if err = database.Create(DBCOLLECTION); err != nil {
			logger.Log("Error while creating collection photos_collection with error : " + err.Error())
			return
		}
		return
	})
	return database, err
}

func createIndexes() (error) {
	db, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return err
	}
	defer db.Close()
	feeds := db.Use(DBCOLLECTION)
	if err := feeds.Index([]string{"Filename"}); err != nil {
		logger.Log("Error while indexing Filename with error : " + err.Error())
		return err
	}
	if err := feeds.Index([]string{"Filepath"}); err != nil {
		logger.Log("Error while indexing Filepath with error : " + err.Error())
		return err
	}
	if err := feeds.Index([]string{"Md5sum","Filepath"}); err != nil {
		logger.Log("Error while indexing Md5sum,Filepath with error : " + err.Error())
		return err
	}
	return nil
}

func InsertNewData(response *modele.PhotoResponse) (error) {
	db, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return err
	}
	defer db.Close()
	feeds := db.Use(DBCOLLECTION)
	for _, item := range response.Photos {
		id, err := feeds.Insert(map[string]interface{}{
			"Filename":item.Filename,
			"Md5sum":item.Md5Sum,
			"Filepath":item.Filepath,
			"ExifTags":item.Tags})
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

func Query(filename string) error {
	md5sumcollection := make([]string, 0)
	dataquery, err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return err
	}
	defer dataquery.Close()
	feeds := dataquery.Use(DBCOLLECTION)
	queryResult := make(map[int]struct{})
	var query interface{}
	json.Unmarshal([]byte(`[{"eq": filename, "in": ["Filename"]}`), &query)

	if err := db.EvalQuery(query, feeds, &queryResult); err != nil {
		logger.Log("Error while querying with error :" + err.Error())
	}
	for id := range queryResult {
		readBack, err := feeds.Read(id)
		if err != nil {
			logger.Log("Error while retreiveing id " + strconv.Itoa(id) + " with error : " + err.Error())
		} else {
			md5sumcollection = append(md5sumcollection, readBack["Md5sum"].(string))
		}

	}

	logger.LogLn(md5sumcollection)
	return nil
}