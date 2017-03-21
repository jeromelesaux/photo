package hash

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"photo/logger"
	"strconv"
)

func Md5Sum(filepath string) (string, error) {
	var err error
	var sum string
	defer func() {
		if err != nil {
			logger.Log(err.Error())
		}
	}()
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	p, err := ioutil.ReadAll(f)
	if err != nil {
		return "", nil
	}
	logger.Log(filepath + "::" + strconv.Itoa(len(p)) + " bytes read.")
	sum = fmt.Sprintf("%x", md5.Sum(p))
	logger.Log(filepath + "::" + sum)
	return sum, nil
}
