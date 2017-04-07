package database

import (
	"photo/modele"
	"strconv"
	"testing"
)

func TestCreateDB(t *testing.T) {
	db, err := NewDataBaseMock()
	if err != nil {
		t.Fatal("database must not be on error with error " + err.Error())
	}
	db.InsertNewData(&modele.PhotoResponse{MachineId: "mymachineid",
		Message: "hello world",
		Version: "1.0",
		Photos:  nil})
	response, _ := db.QueryAll()

	if len(response) != 0 {
		t.Fatal("expected size response 1 and received " + strconv.Itoa(len(response)))
	}
}

func TestInsertAndQueryAll(t *testing.T) {
	db, err := NewDataBaseMock()
	if err != nil {
		t.Fatal("database must not be on error with error " + err.Error())
	}
	photoResponse := &modele.TagsPhoto{Md5Sum: "mymdsum",
		Filepath: "/some/filepath/filename",
		Filename: "filename"}

	db.InsertNewData(&modele.PhotoResponse{MachineId: "mymachineid",
		Message: "hello world",
		Version: "1.0",
		Photos:  []*modele.TagsPhoto{photoResponse}})
	response, _ := db.QueryAll()

	if len(response) != 1 {
		t.Fatal("expected size response 1 and received " + strconv.Itoa(len(response)))
	}
}

func TestGetByName(t *testing.T) {
	db, err := NewDataBaseMock()
	if err != nil {
		t.Fatal("database must not be on error with error " + err.Error())
	}
	photoResponse := &modele.TagsPhoto{Md5Sum: "mymdsum",
		Filepath: "/some/filepath/filename",
		Filename: "filename"}

	db.InsertNewData(&modele.PhotoResponse{MachineId: "mymachineid",
		Message: "hello world",
		Version: "1.0",
		Photos:  []*modele.TagsPhoto{photoResponse}})
	response, _ := db.QueryFilename("file")

	if len(response) != 1 {
		t.Fatal("expected size response 1 and received " + strconv.Itoa(len(response)))
	}
}
