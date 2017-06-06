package folder

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/jeromelesaux/photo/modele"
	"os"
	"path/filepath"
	"strings"
)

func ScanDirectory(r *modele.DirectoryItemResponse) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {

		if err != nil {
			logger.Error(err.Error())
			return err
		}
		if info.IsDir() {
			currentpath := path
			if path[len(path)-1] == '/' {
				currentpath = path[0 : len(path)-1]
			}
			f := filepath.Base(path)
			logger.Info("path scanned " + currentpath)
			//logger.Logf("currenDeep:%d\n",strings.Count(currentpath, "/") )
			current := &modele.DirectoryItemResponse{
				Name:             f,
				Path:             currentpath,
				Directories:      make([]*modele.DirectoryItemResponse, 0),
				Parent:           r,
				JstreeAttributes: modele.NewJSTreeAttribute(),
				Deep:             strings.Count(currentpath, "/")}
			ptr := r
			for {

				if ptr.Deep == 0 || ptr.Deep == (current.Deep-1) {
					ptr.Directories = append(ptr.Directories, current)
					break
				}
				ptr = ptr.Parent
			}
			r = current
		}
		return nil
	}
}
