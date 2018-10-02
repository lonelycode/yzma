package oplog

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/lonelycode/yzma/db"
	"github.com/lonelycode/yzma/logger"
	"github.com/lonelycode/yzma/types/bcaster"
	"github.com/lonelycode/yzma/types/crdt"
	"strconv"
	"strings"
	"time"
)

var log = logger.GetLogger("oplog")
var idGen = crdt.UniqueIDGUIDer{}

type Opn string

const (
	ADD Opn = "ADD"
	REM Opn = "REM"
)

type Replicator interface {
	Send(op *OpLog) error
}

type InAppReplicator struct {
	Buffer chan *OpLog
}

func (r *InAppReplicator) Send(op *OpLog) error {
	if r.Buffer != nil {
		select {
		case r.Buffer <- op:
			// ok
		default:
			// jump
		}
	}

	return nil
}

type PeeringReplicator struct {
	Queue *memberlist.TransmitLimitedQueue
}

func (r *PeeringReplicator) Send(op *OpLog) error {
	msg, err := db.Encode(op)
	if err != nil {
		return err
	}

	if r.Queue != nil {
		r.Queue.QueueBroadcast(&bcaster.Broadcast{Msg: msg, Notify: nil})
	}

	return nil
}

type OpLog struct {
	ID           string        // Operation ID, sortable
	KID          string        // The actual ID that is written Buffer the DB on ADD
	Key          string        // The key used in the interface
	Op           Opn           // The operation (Add, remove etc.
	Value        *crdt.TSValue // What Buffer store
	IsFromRemote bool
}

func NewOp(key string, value interface{}, opn Opn) *OpLog {
	ts := time.Now().UnixNano()
	opId := fmt.Sprintf("%s.%s.%s", string(opn), strconv.Itoa(int(ts)), key)
	vId := idGen.ValueID(nil)
	kId := fmt.Sprintf("%s.%s.%s", strings.ToLower(string(opn)), key, vId)

	return &OpLog{
		ID:    opId,
		KID:   kId,
		Key:   key,
		Op:    opn,
		Value: &crdt.TSValue{TS: ts, Value: value},
	}
}

type Handler struct {
	commitChan  chan *OpLog
	replicaChan chan *OpLog
	db          *db.DB
	rep         Replicator
}

func (h *Handler) SetProcessChannel(ch chan *OpLog) {
	h.commitChan = ch
}

func (h *Handler) SetReplicaChannel(ch chan *OpLog) {
	h.replicaChan = ch
}

func (h *Handler) SetReplicator(rep Replicator) {
	h.rep = rep
}

func (h *Handler) start() {
	for {
		op := <-h.commitChan
		err := h.processOp(op)
		if err != nil {
			log.Error(err)
		}
	}
}

func (h *Handler) startReplChan() {
	for {
		op := <-h.replicaChan
		op.IsFromRemote = true
		err := h.processOp(op)
		if err != nil {
			log.Error(err)
		}
	}
}

func (h *Handler) processOp(op *OpLog) error {
	var err error
	switch op.Op {
	case ADD:
		err = h.db.AddOp(op.KID, op.Value)
	case REM:
		err = h.db.Remove(op.Key)
	default:
		return fmt.Errorf("operation %s not supported", op.Op)
	}

	if err != nil {
		return err
	}

	// don't replicate oplogs from remotes
	if op.IsFromRemote {
		return nil
	}

	return h.replicate(op)
}

func (h *Handler) replicate(op *OpLog) error {
	if h.rep == nil {
		return nil
	}

	err := h.db.StoreOpLog(op.ID, op)
	if err != nil {
		return err
	}

	return h.rep.Send(op)
}

func (h *Handler) Start(db *db.DB) {
	if h.commitChan == nil {
		h.commitChan = make(chan *OpLog)
	}

	if h.db == nil {
		h.db = db
	}

	workers := 1
	for i := 0; i <= workers; i++ {
		go h.start()
		if h.rep != nil {
			go h.startReplChan()
		}
	}

}

func (h *Handler) Add(key string, value interface{}) {
	op := NewOp(key, value, ADD)
	h.commitChan <- op
}

func (h *Handler) Remove(key string) {
	op := NewOp(key, nil, REM)
	h.commitChan <- op
}

func (h *Handler) Replicate(op *OpLog) {
	h.commitChan <- op
}
