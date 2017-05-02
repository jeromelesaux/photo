package exifhandler

import (
	"testing"

	"strconv"
)

func TestExtractThumbnail(t *testing.T) {
	content, err := GetBase64Thumbnail("../vendor/github.com/xiam/exif/_examples/resources/testlocation.jpg")
	if err != nil {
		t.Fatal("Error while trying to extract thumbnail")
	}
	if len(content) != 22324 {
		t.Fatal("Expected size 3464 and get " + strconv.Itoa(len(content)))
	}

}
