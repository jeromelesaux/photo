package database

import (
	"sync"
	"github.com/HouzuoGuo/tiedot/db"
	"photo/logger"
	"photo/modele"
)

var createDB sync.Once
var database *db.DB
var DBCOLLECTION = "photos_collection"

func openDB() (*db.DB, error) {
	var err error
	createDB.Do(func(){

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
     return database,err
}

func InsertNewData(response *modele.PhotoResponse) (error) {
	db,err := openDB()
	if err != nil {
		logger.Log("Cannot use database with error : " + err.Error())
		return err
	}
	defer db.Close()
	feeds := db.Use(DBCOLLECTION)
	for _,item :=range response.Photos {
		id, err := feeds.Insert(map[string]interface{}{
			"Path":item.Filepath,
			"Md5sum":item.Md5Sum,
			"Filepath":item.Filepath,
			"ExifTags":item.Tags})
		if err != nil {
			logger.Log("Cannot insert data in database with error : " +err.Error())
		}else{
			logger.Logf("DB return id %d for filepath:%s\n",id,item.Filepath)
		}

	}


	return nil
}