package database

import (
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"testing"
	"time"
)

func TestSscanf(t *testing.T) {
	var d, m, s float64
	str := "2 deg 14' 32.77\""
	_, err := fmt.Sscanf(str, "%f deg %f' %f", &d, &m, &s)
	if err != nil {
		logger.Errorf("Error while parsing flickr Longitude string %s %v", str, err)
	} else {
		t.Log(Round(d+(m/60)+(s/3600), .5, 2))
	}
}

func TestSscanfWithMinus(t *testing.T) {
	var d, m, s float64
	str := "-2 deg 14' 32.77\""
	_, err := fmt.Sscanf(str, "%f deg %f' %f", &d, &m, &s)
	if err != nil {
		logger.Errorf("Error while parsing flickr Longitude string %s %v", str, err)
	} else {
		t.Log(Round(d+(m/60)+(s/3600), .5, 2))
	}
}

func TestDateConvert(t *testing.T) {
	dateStr := "2015:02:15 15:59:20"
	tm, err := time.Parse("2006:01:02 15:04:05", dateStr)
	if err != nil {
		t.Log(err)
	}
	t.Log(tm)

}

func TestTruncateDateByYear(t *testing.T) {
	dateStr1 := "2015:02:15 15:59:20"
	dateStr2 := "2015:04:15 15:59:20"
	tm1, err := time.Parse("2006:01:02 15:04:05", dateStr1)
	if err != nil {
		t.Log(err)
	}
	tm2, err := time.Parse("2006:01:02 15:04:05", dateStr2)
	if err != nil {
		t.Log(err)
	}
	rounded1 := time.Date(tm1.Year(), 0, 0, 0, 0, 0, 0, tm1.Location())
	rounded2 := time.Date(tm2.Year(), 0, 0, 0, 0, 0, 0, tm2.Location())
	if rounded1 != rounded2 {
		t.Fatal("expected same rounded date and failed")
	}
}
