package peering

import (
	"github.com/lonelycode/yzma/logger"
	"github.com/lonelycode/yzma/oplog"
	"github.com/spf13/viper"
)

type PeerData struct {
	NodeName   string
	APIIngress string
	Token      string
}

type PeerConfig struct {
	Name             string
	AdvertiseAddress string
	AdvertisePort    int
	BindPort         int
	BindAddr         string
	Join             []string
	Block            []string
	Federation       *PeerData
	ReplicaChan      chan *oplog.OpLog
	OpLogHandler     *oplog.Handler
}

type Config struct {
	Peering *PeerConfig
}

var sconf = &Config{}

var moduleName = "peering"
var log = logger.GetLogger(moduleName)

func GetConf() *PeerConfig {
	err := viper.Unmarshal(sconf)
	if err != nil {
		log.Fatal("failed to read peering config: ", err)
	}

	return sconf.Peering
}
