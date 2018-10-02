package main

import (
	"fmt"
	"github.com/lonelycode/yzma/logger"
	"github.com/lonelycode/yzma/peering"
	"github.com/lonelycode/yzma/server"
	"github.com/satori/go.uuid"
	"os"
	"os/signal"
	"sync"
)

var log = logger.GetLogger("yzma")

func main() {
	svr := server.Server{}
	svr.Start(uuid.NewV4().String())

	// start the peer manager
	peering.Start(nil)

	// Wait to quit
	log.Info("Press Ctrl+C to end")
	waitForCtrlC()
	fmt.Printf("\n")

	// End gracefully
	err := peering.Stop()
	if err != nil {
		log.Error(err)
	}
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
