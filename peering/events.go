package peering

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"strings"
)

type Definition struct {
	Tag        string
	ListenPort string
	ListenAddr string
}

func (d *Definition) Addr() string {
	addr := fmt.Sprintf("%v:%v", d.ListenAddr, d.ListenPort)
	return addr
}

type PeerEvents struct {
	blockList []string
	servList  []*Definition
}

func (p *PeerEvents) isBlocked(n string) bool {
	for _, b := range p.blockList {
		if strings.ToLower(b) == strings.ToLower(n) {
			return true
		}
	}

	return false
}

func (p *PeerEvents) NotifyJoin(n *memberlist.Node) {
	log.Info("join event from ", n.Name)
	remoteCfg := &PeerData{}
	err := json.Unmarshal(n.Meta, remoteCfg)
	if err != nil {
		log.Error("failed to read node metadata: ", err)
	}

	if p.isBlocked(remoteCfg.NodeName) {
		log.Error("peer is blocked, ignoring: ", remoteCfg.NodeName)
		return
	}

	for _, svr := range p.servList {
		if svr.Tag == remoteCfg.NodeName {
			log.Info("detected existing server, skipping")
			return
		}
	}

	log.WithFields(logrus.Fields{
		"Node": n.Name,
		"Addr": n.Address(),
	}).Info("node joined: ", n.Name)

}

func (p *PeerEvents) NotifyLeave(n *memberlist.Node) {
	remoteCfg := &PeerData{}
	err := json.Unmarshal(n.Meta, remoteCfg)
	if err != nil {
		log.Error("could not decode node data: ", n.Meta)
	}

	log.WithFields(logrus.Fields{
		"Node": n.Name,
		"Addr": n.Address(),
	}).Info("node left: ", n.Name)
}

func (p *PeerEvents) NotifyUpdate(n *memberlist.Node) {
	log.WithFields(logrus.Fields{
		"Node": n.Name,
		"Addr": n.Address(),
	}).Info("node updating: ", n.Name)
}

func (p *PeerManager) Init(cfg *PeerConfig) error {
	p.cfg = cfg

	p.Broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return p.members.NumMembers()
		},
		RetransmitMult: 3,
	}

	listCfg := memberlist.DefaultWANConfig()
	listCfg.Name = p.cfg.Name
	listCfg.AdvertiseAddr = p.cfg.AdvertiseAddress
	listCfg.AdvertisePort = p.cfg.AdvertisePort
	listCfg.BindPort = p.cfg.BindPort
	listCfg.Events = &PeerEvents{blockList: p.cfg.Block, servList: p.fixedServers}
	listCfg.Delegate = &PeerDelegate{
		cfg:          p.cfg.Federation,
		bcast:        p.Broadcasts,
		bcastChan:    p.cfg.ReplicaChan,
		oplogHandler: p.cfg.OpLogHandler,
	}
	listCfg.BindAddr = p.cfg.BindAddr

	bAddr := "0.0.0.0"
	if p.cfg.BindAddr != "" {
		bAddr = p.cfg.BindAddr
	}

	webTSConf := &WebTransportConfig{
		BindAddrs: []string{bAddr},
		BindPort:  p.cfg.BindPort,
	}

	var err error
	listCfg.Transport, err = NewWebTransport(webTSConf)
	if err != nil {
		return err
	}

	list, err := memberlist.Create(listCfg)
	if err != nil {
		return fmt.Errorf("Failed to create memberlist: " + err.Error())
	}

	log.Info("peer-list binding to ", list.LocalNode().Addr.String())

	p.members = list

	if len(p.cfg.Join) > 0 {
		addrList, err := resolveList(p.cfg.Join)
		if err != nil {
			return err
		}

		log.Info("detected peer list, attempting to join")
		jn, err := p.members.Join(addrList)
		if err != nil {
			log.Error("failed to join peers")
			// don't return error, this just means the peers are down
			return nil
		}

		log.Info("joined ", jn, " of ", len(p.cfg.Join), " peers")
	}

	return nil
}
