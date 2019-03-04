package server

import (
	"github.com/lonelycode/yzma/oplog"
	"github.com/lonelycode/yzma/peering"
	"github.com/lonelycode/yzma/types/crdt"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"
)

var jsObj = `
{
	"glossary": {
		"title": "example glossary",
		"GlossDiv": {
			"title": "S",
			"GlossList": {
				"GlossEntry": {
					"ID": "SGML",
					"SortAs": "SGML",
					"GlossTerm": "Standard Generalized Markup Language",
					"Acronym": "SGML",
					"Abbrev": "ISO 8879:1986",
					"GlossDef": {
						"para": "A meta-markup language, used to create markup languages such as DocBook.",
						"GlossSeeAlso": ["GML", "XML"]
					},
					"GlossSee": "markup"
				}
			}
		}
	}
}
`

func GetVal(v crdt.Payload) interface{} {
	d, _ := v.Extract()
	return d
}

func TestServerAndReplication(t *testing.T) {
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
	s1.Add("k1", []byte("foo"), "")
	s1.Add("k2", []byte("bar"), "")
	s1.Add("k3", []byte("baz"), "")
	s1.Add("k4", []byte(jsObj), "")

	time.Sleep(100 * time.Millisecond)
	v, ok := s1.Load("k1")
	if !ok {
		t.Error("expected to find k1")
	}
	if string(GetVal(v).([]byte)) != "foo" {
		t.Error("wrong value for k1: ", string(GetVal(v).([]byte)))
	}

	v, ok = s1.Load("k2")
	if !ok {
		t.Error("expected to find k2")
	}
	if string(GetVal(v).([]byte)) != "bar" {
		t.Error("wrong value for k2", string(GetVal(v).([]byte)))
	}

	v, ok = s1.Load("k3")
	if !ok {
		t.Error("expected to find k3")
	}
	if string(GetVal(v).([]byte)) != "baz" {
		t.Error("wrong value for k3", string(GetVal(v).([]byte)))
	}

	v, ok = s1.Load("k4")
	if !ok {
		t.Error("expected to find k4")
	}
	if string(GetVal(v).([]byte)) != jsObj {
		t.Error("wrong value for k4", string(GetVal(v).([]byte)))
	}

	s2.Add("k1", []byte("barbaz"), "")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		time.Sleep(10 * time.Second)
		v2, _ := s1.Load("k1")
		d, _ := v2.Extract()
		if string(d.([]byte)) != "barbaz" {
			t.Error("expected s1 to have replicated k1 from s2")
		}
		wg.Done()

	}()

	wg.Wait()

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
