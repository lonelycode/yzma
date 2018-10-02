package server

import (
	"github.com/lonelycode/yzma/db"
	"github.com/lonelycode/yzma/logger"
	"github.com/lonelycode/yzma/oplog"
	"github.com/lonelycode/yzma/peering"
	"github.com/lonelycode/yzma/types/crdt"
)

type Config struct {
	DBPath string
}

type Server struct {
	opHandler *oplog.Handler
	db        *db.DB
	cfg       *Config
	peers     *peering.PeerManager
}

var log = logger.GetLogger("server")

func (s *Server) Start(name string, peeringCfg *peering.PeerConfig, stopCh chan struct{}) {
	// Create and start a peer handler
	pm, err := peering.NewPeerManager(peeringCfg)
	if err != nil {
		log.Fatal(err)
	}
	s.peers = pm

	// Create a DB
	d, err := db.New(s.cfg.DBPath)
	if err != nil {
		panic(err)
	}
	d.Options.CollisionStrategy = crdt.LWWStrat
	s.db = d

	// Create an OpHandler
	s.opHandler = &oplog.Handler{}
	s.opHandler.SetReplicaChannel(peeringCfg.ReplicaChan)
	// Create a replicator
	s.opHandler.SetReplicator(&oplog.PeeringReplicator{
		Queue: s.peers.Broadcasts,
	})

	s.opHandler.Start(d)

	<-stopCh
	log.Warn("received stop signal, exiting")

}

func (s *Server) Stop() {
	// End gracefully
	err := s.peers.Leave()
	if err != nil {
		log.Error(err)
	}
}

func (s *Server) Add(key string, value interface{}) {
	s.opHandler.Add(key, value)
}

func (s *Server) Remove(key string, value interface{}) {
	s.opHandler.Remove(key)
}

func (s *Server) Load(key string) (crdt.Payload, bool) {
	return s.db.Load(key)
}
