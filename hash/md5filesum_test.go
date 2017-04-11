package hash

import (
	"testing"
)

func TestMd5Sum(t *testing.T) {
	hash, err := Md5Sum("md5filesum.go")
	if err != nil {
		t.Fatal("err must nont fired with error message " + err.Error())
	}
	if hash != "d37434a88a17d2f52df9fbfb6b7692b1" {
		t.Fatal("hash must have value 1105710f16a1576d1f03e8d329ce422e and return " + hash)
	}

}
