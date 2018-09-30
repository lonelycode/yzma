package crdt

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"sort"
	"strings"
	"time"
)

const (
	LWWStrat = "lww"
	NoStrat = ""
)

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

type ORSet struct {
	IDSource ObserveGUIDer
	Options struct {
		CollisionStrategy string
	}
	addMap map[interface{}]map[string]*TSValue
	rmMap  map[interface{}]map[string]*TSValue
}

type TSValue struct {
	TS int64
	Value interface{}
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

func NewORSet() *ORSet {
	return &ORSet{
		addMap: make(map[interface{}]map[string]*TSValue),
		rmMap:  make(map[interface{}]map[string]*TSValue),
		IDSource: &UniqueIDGUIDer{},
	}
}

func (o *ORSet) AddOrUpdate(key, value interface{}) {
	// Guarantees a delete and Add for every value
	_, ok := o.Load(key)
	if !ok {
		o.Add(key, value)
		return
	}

	o.Remove(key)
	o.Add(key, value)
}

func (o *ORSet) Add(key, value interface{}) {
	tsVal := &TSValue{
		TS: time.Now().UnixNano(),
		Value: value,
	}

	vId := o.IDSource.ValueID(value)
	if m, ok := o.addMap[key]; ok {
		m[vId] = tsVal
		o.addMap[key] = m
		return
	}

	m := make(map[string]*TSValue)

	m[vId] = tsVal
	o.addMap[key] = m
}

func (o *ORSet) Remove(key interface{}) {
	r, ok := o.rmMap[key]
	if !ok {
		r = make(map[string]*TSValue)
	}

	if m, ok := o.addMap[key]; ok {
		for uid, v := range m {
			r[uid] = v
		}
	}

	o.rmMap[key] = r
}

func (o *ORSet) HandleCollision(values Payload) Payload {

	if o.Options.CollisionStrategy == LWWStrat {
		times := make([]int, len(values))
		i := 0
		for _, v := range values {
			tsv := v
			times[i] = int(tsv.TS)
			i++
		}

		sort.Ints(times)
		last := times[len(times)-1]
		for id, v := range values {
			tsv := v
			if int(tsv.TS) == last {
				return map[string]*TSValue{id:v}
			}
		}
	}

	return values
}

func (o *ORSet) Load(key interface{}) (Payload, bool) {
	// If never added, then return false
	addMap, ok := o.addMap[key]
	if !ok {
		return nil, false
	}

	// it has been added, has it ever been removed? If not, then it exists
	rmMap, ok := o.rmMap[key]
	if !ok {
		return o.HandleCollision(addMap), true
	}

	// It has been added, and it has been removed at some point
	retMap := make(map[string]*TSValue)
	for uid, v := range addMap {
		if _, ok := rmMap[uid]; !ok {
			retMap[uid] = v
		}
	}

	if len(retMap) > 0 {
		return o.HandleCollision(retMap), true
	}

	return nil, false
}

func (o *ORSet) Unique() []interface{} {
	fin := make([]interface{},0)
	// Only use non-removed values
	for k, uids := range o.addMap  {
		rmMap, ok := o.rmMap[k]

		// It has never been changed or deleted, so we can assume it is a safe value
		if !ok {
			fin = append(fin, k)
			continue
		}

		for uid := range uids {
			if _, ok := rmMap[uid]; !ok {
				fin = append(fin, k)
			}
		}
	}

	return fin
}

func (o *ORSet) Merge(r *ORSet) {
	for value, m := range r.addMap {
		addMap, ok := o.addMap[value]
		if ok {
			for uid, v := range m {
				addMap[uid] = v
			}

			continue
		}

		o.addMap[value] = m
	}

	for value, m := range r.rmMap {
		rmMap, ok := o.rmMap[value]
		if ok {
			for uid, v := range m {
				rmMap[uid] = v
			}

			continue
		}

		o.rmMap[value] = m
	}
}
