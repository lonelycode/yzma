package api

import "github.com/spf13/viper"

type APICfg struct {
	Bind string
}

type Config struct {
	API *APICfg
}

var sconf = &Config{}

func GetConf() *APICfg {
	err := viper.Unmarshal(sconf)
	if err != nil {
		log.Fatal("failed to read Db config: ", err)
	}

	return sconf.API
}
