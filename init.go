package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
)

var cfgPtr, joinPtr *string // late-binding for configuration location

func ReadConfig() {
	err := viper.ReadInConfig()
	if err != nil {
		if os.Getenv("TESTING") == "" {
			log.Fatal(fmt.Errorf("fatal error config file: %s \n", err))
		}
	}
}

func init() {
	// support ara.* files
	viper.SetConfigName("yzma")

	// Three main locations to look for config files
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.yzma")
	viper.AddConfigPath("/etc/yzma/")

	// Allow for a custom location to be set
	cfgPtr = flag.String("c", "", "config file")
	joinPtr = flag.String("j", "", "join peer")
	flag.Parse()

	if *cfgPtr != "" {
		viper.SetConfigFile(*cfgPtr)
	}

	ReadConfig()
}
