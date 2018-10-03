package server

import (
	"github.com/lonelycode/yzma/db"
	"github.com/lonelycode/yzma/logger"
	"github.com/lonelycode/yzma/oplog"
	"github.com/lonelycode/yzma/peering"
	"github.com/lonelycode/yzma/types/crdt"
)

type Server struct {
	opHandler *oplog.Handler
	db        *db.DB
	cfg       *Config
	peers     *peering.PeerManager
	stopCh    chan struct{}
}

var log = logger.GetLogger("server")

func (s *Server) Start(name string, peeringCfg *peering.PeerConfig, stopCh chan struct{}) {
	s.stopCh = stopCh

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

	log.Info("starting oplog processor")
	s.opHandler.Start(d)
	log.Info("db ready")

	<-stopCh
	log.Warn("received stop signal, stopping service")
	s.opHandler.Stop()

}

func (s *Server) Stop() {
	// End gracefully
	log.Info("leaving peer group")
	err := s.peers.Leave()
	if err != nil {
		log.Error(err)
	}

	log.Info("closing DB")
	s.db.Close()

	log.Info("stopping server")
	s.stopCh <- struct{}{}
}

func (s *Server) Add(key string, value interface{}) {
	s.opHandler.Add(key, value)
}

func (s *Server) Remove(key string) {
	s.opHandler.Remove(key)
}

func (s *Server) Load(key string) (crdt.Payload, bool) {
	return s.db.Load(key)
}

func (s *Server) Join(peers []string) error {
	return s.peers.Join(peers)
}

func (s *Server) Leave() error {
	return s.peers.Leave()
}

func (s *Server) SetConfig(cfg *Config) {
	s.cfg = cfg
}
