package config

import (
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
)

var (
	AppPath string
	AppName string
	AppMode string

	defaultConf *ini.File
)

func SetDefault(file string) {
	var err error
	if AppPath, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
		panic(err)
	}

	defaultConf = NewConfig(file)
	AppName = Key("name").MustString("app")
	AppMode = Key("mode").MustString("release")
}

func NewConfig(configFile string) *ini.File {
	//var configPath string
	//configPath = filepath.Join(AppPath, "config", filename)
	//if !utils.FileExists(configPath) {
	//	tempPath, err := os.Getwd()
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	configPath = filepath.Join(tempPath, "config", filename)
	//	for !utils.FileExists(configPath) {
	//		if tempPath == "" {
	//			log.Println(fmt.Sprintf("config file %s not existed!", filename))
	//			return nil
	//		}
	//		tempPath = utils.ParentDirectory(tempPath)
	//		configPath = filepath.Join(tempPath, "config", filename)
	//	}
	//}

	cfg, err := ini.Load(configFile)
	if err != nil {
		panic(err)
	}
	return cfg
}

func Section(name string) *ini.Section {
	if defaultConf == nil {
		panic("default conf not found.")
	}
	return defaultConf.Section(name)
}

func Key(name string) *ini.Key {
	return Section("app").Key(name)
}
