package peering

import (
	"errors"
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/satori/go.uuid"
	"net"
	"strings"
	"time"
)

type update struct {
	Action string // add, del
	Data   map[string]string
}

type PeerManager struct {
	cfg          *PeerConfig
	members      *memberlist.Memberlist
	fixedServers []*Definition
	Broadcasts   *memberlist.TransmitLimitedQueue
	Name         string
}

func (p *PeerManager) Join(peers []string) error {
	_, err := p.members.Join(peers)

	return err
}

func (p *PeerManager) Leave() error {
	log.Info("received leave request")
	err := p.members.Leave(time.Second * 30)
	if err != nil {
		return err
	}

	return nil
}

func resolveList(hosts []string) ([]string, error) {
	out := make([]string, len(hosts))

	for i, addr := range hosts {
		parts := strings.Split(addr, ":")
		if len(parts) != 2 {
			return nil, errors.New("address had more than two parts")
		}

		ipAddr, err := net.ResolveIPAddr("ip", parts[0])
		if err != nil {
			return nil, err
		}

		newAddr := fmt.Sprintf("%v:%v", ipAddr.String(), parts[1])
		out[i] = newAddr
	}

	return out, nil
}

func NewPeerManager(cfg *PeerConfig) (*PeerManager, error) {
	// Guarantee uniqueness because we may have
	// more than one container
	cfg.Name = cfg.Name + "-" + uuid.NewV4().String()

	pm := &PeerManager{}

	// Reference of what's in our config
	err := pm.Init(cfg)
	if err != nil {
		return nil, err
	}

	return pm, nil
}
