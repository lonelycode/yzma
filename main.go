package main

import (
	"fmt"
	"github.com/lonelycode/yzma/api"
	"github.com/lonelycode/yzma/logger"
	"github.com/lonelycode/yzma/oplog"
	"github.com/lonelycode/yzma/peering"
	"github.com/lonelycode/yzma/server"
	"os"
	"os/signal"
	"sync"
	"time"
)

var log = logger.GetLogger("yzma")

func info() {
	msg := fmt.Sprintf("\nYzmaDB %v copyright Martin Buhr %v", Version, time.Now().Year())
	line := ""
	for i := 0; i < len(msg)-1; i++ {
		line += "="
	}

	fmt.Println(msg)
	fmt.Println(line)
	fmt.Println()
}

func doJoin(svr *server.Server) {
	for {
		if svr.Ready() {
			err := svr.Join([]string{*joinPtr})
			if err != nil {
				log.Error("join failed: ", err)
			}
			break
		}

		time.Sleep(1 * time.Second)
	}
}

func main() {
	info()
	peeringConf := peering.GetConf()
	peeringConf.ReplicaChan = make(chan *oplog.OpLog)
	svrConf := server.GetConf()

	stopChan := make(chan struct{})
	svr := &server.Server{}
	svr.SetConfig(svrConf)
	go svr.Start("dat.db", peeringConf, stopChan)

	webCfg := api.GetConf()
	web := api.WebAPI{}
	go web.Start(svr, webCfg)

	if *joinPtr != "" {
		go doJoin(svr)
	}

	// Wait to quit
	log.Info("Press Ctrl+C to end")
	waitForCtrlC()
	fmt.Printf("\n")
	log.Info("exiting...")

	svr.Stop()
}

func waitForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	var sigChannel chan os.Signal
	sigChannel = make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)
	go func() {
		<-sigChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}
