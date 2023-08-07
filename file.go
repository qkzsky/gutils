package gutils

import (
	"os"
	"strings"
)

func FileExists(name string) (bool, error) {
	if _, err := os.Stat(name); err != nil {
		//if !os.IsNotExist(err) {
		//	return error
		//	//logger.Error(err.Error())
		//}
		return false, err
	}
	return true, nil
}

func ParentDirectory(directory string) string {
	return SubStr(directory, 0, strings.LastIndex(directory, string(os.PathSeparator)))
}
