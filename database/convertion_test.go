package database

import (
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"testing"
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
