package server

import "github.com/spf13/viper"

type Config struct {
	DBPath string
}

type MainConfig struct {
	Server *Config
}

var sconf = &MainConfig{}

func GetConf() *Config {
	err := viper.Unmarshal(sconf)
	if err != nil {
		log.Fatal("failed to read peering config: ", err)
	}

	return sconf.Server
}
