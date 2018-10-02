package server

import (
	"github.com/lonelycode/yzma/oplog"
	"github.com/lonelycode/yzma/peering"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	replicaChanS1 := make(chan *oplog.OpLog)
	replicaChanS2 := make(chan *oplog.OpLog)

	pConfS1 := &peering.PeerConfig{
		Name:             "s1",
		BindPort:         37001,
		BindAddr:         "0.0.0.0",
		AdvertisePort:    37001,
		AdvertiseAddress: "127.0.0.1",
		Federation: &peering.PeerData{
			NodeName:   "s1",
			APIIngress: "127.0.0.1:37002",
			Token:      "foo",
		},
		ReplicaChan: replicaChanS1,
	}

	pConfS2 := &peering.PeerConfig{
		Name:             "s2",
		BindPort:         38001,
		BindAddr:         "0.0.0.0",
		AdvertisePort:    38001,
		AdvertiseAddress: "127.0.0.1",
		Federation: &peering.PeerData{
			NodeName:   "s2",
			APIIngress: "127.0.0.1:38002",
			Token:      "foo",
		},
		ReplicaChan: replicaChanS2,
	}

	s1 := Server{
		cfg: &Config{
			DBPath: "s1.db",
		},
	}
	defer os.Remove("s1.db")
	s1Stop := make(chan struct{})
	go s1.Start("s1", pConfS1, s1Stop)

	s2 := Server{
		cfg: &Config{
			DBPath: "s2.db",
		},
	}
	defer os.Remove("s2.db")

	s2Stop := make(chan struct{})
	go s2.Start("s2", pConfS2, s2Stop)

	time.Sleep(1 * time.Second)
	s2.peers.Join([]string{"127.0.0.1:37001"})

	time.Sleep(100 * time.Millisecond)
	s1.Add("foo", "bar")

	time.Sleep(100 * time.Millisecond)
	v, ok := s1.Load("foo")
	log.Info("S1: ", ok, v.Extract())

	s2.Add("foo", "barbaz")

	go func() {
		time.Sleep(10 * time.Second)
		v2, _ := s1.Load("foo")
		log.Info("S1: ", v2.Extract())

		v3, _ := s2.Load("foo")
		log.Info("S2: ", v3.Extract())
	}()

	waitForCtrlC()
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
