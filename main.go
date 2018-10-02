package main

import (
	"fmt"
	"github.com/lonelycode/yzma/logger"
	"os"
	"os/signal"
	"sync"
)

var log = logger.GetLogger("yzma")

func main() {

	// Wait to quit
	log.Info("Press Ctrl+C to end")
	waitForCtrlC()
	fmt.Printf("\n")

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
