package exifhandler

import (
	"strconv"
	"testing"
)

func TestExtractThumbnail(t *testing.T) {
	content, err := GetBase64Thumbnail("../vendor/github.com/xiam/exif/_examples/resources/testlocation.jpg")
	if err != nil {
		t.Fatal("Error while trying to extract thumbnail")
	}
	if len(content) != 22300 {
		t.Fatal("Expected size 22300 and get " + strconv.Itoa(len(content)))
	}

}
