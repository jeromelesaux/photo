package routes

import (
	"github.com/Sirupsen/logrus"
	"net/url"
	"strings"
	"testing"
)

func TestParsingArrayUrl(t *testing.T) {
	u, err := url.Parse("pdfalbum?albumName=jpg&photosid=\"20659910082\",\"20481128950\",\"20481135158\"")
	if err != nil {
		logrus.Fatalf("Error with %v", err)
	}
	c, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		logrus.Fatalf("Error with %v", err)
	}
	nc := strings.Split(c["photosid"][0], ",")

	if len(nc) != 3 {
		t.Fatalf("expected 2 and received :%d", len(c["photosid"]))
	}
}
