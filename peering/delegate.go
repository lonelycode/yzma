package peering

import (
	"encoding/json"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/memberlist"
	"github.com/lonelycode/yzma/db"
	"github.com/lonelycode/yzma/oplog"
	"sync"
)

type PeerDelegate struct {
	cfg          *PeerData
	bcast        *memberlist.TransmitLimitedQueue
	bcastChan    chan *oplog.OpLog
	oplogHandler *oplog.Handler
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

	log.Debug("Message is: ", string(b))
	op := &oplog.OpLog{}
	err := db.Decode(b, op)
	if err != nil {
		log.Error(err)
	}

	if p.bcastChan != nil {
		select {
		case p.bcastChan <- op:
			log.Debug("notification sent to DB")
		default:
			log.Debug("notification bounced, channel busy")
		}
	}
	return
}

func (p *PeerDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return p.bcast.GetBroadcasts(overhead, limit)
}

func (p *PeerDelegate) LocalState(join bool) []byte {
	if !join {
		return nil
	}

	olog := p.oplogHandler.OpLog("")
	var dat []byte
	h := &codec.MsgpackHandle{}
	var enc = codec.NewEncoderBytes(&dat, h)
	err := enc.Encode(olog)
	if err != nil {
		log.Error(err)
		return nil
	}

	return dat
}

func (p *PeerDelegate) MergeRemoteState(buf []byte, join bool) {
	log.Warning("merge remote state called")
	if len(buf) == 0 {
		return
	}
	if !join {
		log.Debug("not a join, returning")
		return
	}

	var h codec.Handle = new(codec.MsgpackHandle)
	var dec = codec.NewDecoderBytes(buf, h)
	var arrayDat [][]byte
	err := dec.Decode(&arrayDat)
	if err != nil {
		log.Error(err)
	}

	// TODO: This could probably be *much* faster
	// by reading and writing direct to the DB
	for _, bOp := range arrayDat {
		opVal := &oplog.OpLog{}
		err = json.Unmarshal(bOp, opVal)
		if err != nil {
			log.Error(err)
			continue
		}

		p.oplogHandler.Replicate(opVal)
	}

}
