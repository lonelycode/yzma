package crdt

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/lonelycode/yzma/logger"
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

const (
	LWWStrat = "lww"
	NoStrat = ""
)

var log = logger.GetLogger("crdt")


type ObserveGUIDer interface{
	ValueID(value interface{}) string
}

type ValueHashGUIDer struct{}
func (v ValueHashGUIDer) ValueID(value interface{}) string {
	s, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	md5 := md5.Sum([]byte(s))
	return fmt.Sprintf("%x\n", md5)
}

type UniqueIDGUIDer struct {}
func (v UniqueIDGUIDer) ValueID(value interface{}) string {
	id := strings.Replace(uuid.NewV4().String(), "-", "", -1)
	tsStr := fmt.Sprintf("%v", time.Now().UnixNano())
	return fmt.Sprintf("%s:%s", id, tsStr)
}



type TSValue struct {
	TS int64
	Value []byte
}

type Payload map[string]*TSValue

func (p Payload) Extract() interface{} {
	if len(p) == 1 {
		for _, v := range p {
			return v.Value
		}
	}

	return nil
}

func (p Payload) ExtractAll() []interface{} {
	rets := make([]interface{}, len(p))
	i := 0
	for _, v := range p {
		rets[i] = v
		i++
	}

	return rets
}