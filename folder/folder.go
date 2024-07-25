package folder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jeromelesaux/photo/modele"
	logger "github.com/sirupsen/logrus"
)

func ScanSubDirectory(r *modele.DirectoryItemResponse, directory string) (*modele.DirectoryItemResponse, error) {
	d, err := os.Open(directory)
	if err != nil {
		logger.Errorf("Error while scanning directory %s with error %v", directory, err)
		return r, err
	}
	defer d.Close()
	fi, err := d.Readdir(-1)
	if err != nil {
		logger.Errorf("Error while scanning directory %s with error %v", directory, err)
		return r, err
	}
	for _, fi := range fi {
		if fi.IsDir() {
			currentpath := directory + "/" + fi.Name()
			current := &modele.DirectoryItemResponse{
				Name:             fi.Name(),
				Path:             currentpath,
				Directories:      make([]*modele.DirectoryItemResponse, 0),
				Parent:           r,
				JstreeAttributes: modele.NewJSTreeAttribute(),
				Deep:             strings.Count(currentpath, "/")}

			r.Directories = append(r.Directories, current)

		}
	}
	return r, nil
}

func ScanDirectory(r *modele.DirectoryItemResponse) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {

		if err != nil {
			logger.Error(err.Error())
			return err
		}
		if info.IsDir() {
			//if r.Stop() {
			//	return nil
			//}
			currentpath := path
			if path[len(path)-1] == '/' {
				currentpath = path[0 : len(path)-1]
			}
			f := filepath.Base(path)
			logger.Info("path scanned " + currentpath)
			logger.Infof("currenDeep:%d\n", strings.Count(currentpath, "/"))
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
