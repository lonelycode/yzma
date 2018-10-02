package db

import (
	"github.com/lonelycode/yzma/logger"
	"github.com/spf13/viper"
)

type ModuleConf struct {
	FileName string
}

type Config struct {
	DB *ModuleConf
}

var sconf = &Config{}

var moduleName = "Db"
var log = logger.GetLogger(moduleName)

func GetConf() *ModuleConf {
	err := viper.Unmarshal(sconf)
	if err != nil {
		log.Fatal("failed to read Db config: ", err)
	}

	return sconf.DB
}