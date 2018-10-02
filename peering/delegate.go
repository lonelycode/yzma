package peering

import (
	"encoding/json"
	"github.com/hashicorp/memberlist"
	"github.com/lonelycode/yzma/db"
	"github.com/lonelycode/yzma/oplog"
	"sync"
)

type PeerDelegate struct {
	cfg       *PeerData
	bcast     *memberlist.TransmitLimitedQueue
	bcastChan chan *oplog.OpLog
}

var (
	mtx   sync.RWMutex
	items = map[string]string{}
)

func (p *PeerDelegate) NodeMeta(limit int) []byte {
	js, err := json.Marshal(p.cfg)
	if err != nil {
		log.Error(err)
	}

	log.Debug("metadata returned: ", string(js))
	return js
}

func (p *PeerDelegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	log.Info("Message is: ", string(b))
	op := &oplog.OpLog{}
	err := db.Decode(b, op)
	if err != nil {
		log.Error(err)
	}

	if p.bcastChan != nil {
		select {
		case p.bcastChan <- op:
			log.Info("notification sent to DB")
		default:
			log.Warning("notification bounced, channel busy")
		}
	}
	return
}

func (p *PeerDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return p.bcast.GetBroadcasts(overhead, limit)
}

func (p *PeerDelegate) LocalState(join bool) []byte {
	mtx.RLock()
	m := items
	mtx.RUnlock()
	b, _ := json.Marshal(m)
	return b
}

func (p *PeerDelegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	mtx.Lock()
	for k, v := range m {
		items[k] = v
	}
	mtx.Unlock()
}
