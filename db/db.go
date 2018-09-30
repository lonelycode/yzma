package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/lonelycode/yzma/types/crdt"
	bolt "go.etcd.io/bbolt"
	"sort"
	"strings"
	"sync"
	"time"
)

type DB struct {
	db *bolt.DB
	IDSource crdt.ObserveGUIDer
	Options struct {
		CollisionStrategy string
	}
}

const (
	KEYS = "keys"
)

var ReadyDBs = sync.Map{}

func New(file string) (*DB, error) {
	ex, ok := ReadyDBs.Load(file)
	if ok {
		return ex.(*DB), nil
	}

	db := &DB{}
	err := db.Init(file)
	if err != nil {
		return nil, err
	}

	db.IDSource = &crdt.UniqueIDGUIDer{}

	ReadyDBs.Store(file, db)
	return db, nil
}

func (d *DB) Init(path string) error {
	db, err := bolt.Open(path, 0666, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	d.db = db

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(KEYS))
		if err != nil {
			return fmt.Errorf("create addMap bucket: %s", err)
		}

		return nil
	})

	return nil
}

func (d *DB) Close() {
	d.db.Close()
}

func (d *DB) Add(key string, value interface{}) error {
	vId := d.IDSource.ValueID(value)
	tsv := &crdt.TSValue{TS: time.Now().UnixNano(), Value: value}
	addKey := fmt.Sprintf("add.%s.%s", key, vId)

	enc, err := Encode(tsv)
	if err != nil {
		return err
	}

	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(KEYS))
		b.Put([]byte(addKey), enc)
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *DB) Remove(key string) error {

	// we must copy IDs over for anything already added
	rmMap := map[string]*crdt.TSValue{}
	d.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(KEYS)).Cursor()
		addPrefix := []byte(fmt.Sprintf("add.%s", key))
		for k, _ := c.Seek(addPrefix); k != nil && bytes.HasPrefix(k, addPrefix); k, _ = c.Next() {
			rmMap[string(k)] = &crdt.TSValue{}
		}

		return nil
	})

	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(KEYS))

		for k  := range rmMap {
			remKey := strings.Replace(k, "add.", "rem.", 1)
			b.Put([]byte(remKey), []byte{})
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *DB) Load(key string) (crdt.Payload, bool) {
	addPrefix := []byte(fmt.Sprintf("add.%s", key))
	var retPL crdt.Payload
	var found bool
	d.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(KEYS)).Cursor()

		addMap := map[string]*crdt.TSValue{}
		for k, v := c.Seek(addPrefix); k != nil && bytes.HasPrefix(k, addPrefix); k, v = c.Next() {
			tsv := &crdt.TSValue{}
			if err := Decode(v, tsv); err != nil {
				return err
			}

			addMap[d.GetUIDFromKey(string(k))] = tsv
		}

		// Never added, so not found
		if len(addMap) == 0 {
			found = false
			return nil
		}

		remPrefix := []byte(fmt.Sprintf("rem.%s", key))
		rmMap := map[string]*crdt.TSValue{}
		for k, _ := c.Seek(remPrefix); k != nil && bytes.HasPrefix(k, remPrefix); k, _ = c.Next() {
			rmMap[d.GetUIDFromKey(string(k))] = &crdt.TSValue{}
		}

		// never removed, so found
		if len(rmMap) == 0 {
			found = true
			retPL = addMap
			return nil
		}

		// It has been added, and it has been removed at some point
		retPL = crdt.Payload{}
		for id, v := range addMap {
			uid := d.GetUIDFromKey(id)
			if _, ok := rmMap[uid]; !ok {
				retPL[uid] = v
			}
		}

		if len(retPL) == 0 {
			found = false
			return nil
		}

		found = true
		return nil
	})

	if len(retPL) == 0 {
		return nil, false
	}

	return d.HandleCollision(retPL), found
}

func (d *DB) GetUIDFromKey(k string) string {
	pts := strings.Split(k, ".")
	if len(pts) != 3 {
		return k
	}
	uid := pts[2]

	return uid
}

func (d *DB) HandleCollision(values crdt.Payload) crdt.Payload {
	if d.Options.CollisionStrategy == crdt.LWWStrat {
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
				ret := map[string]*crdt.TSValue{id:v}
				return ret
			}
		}
	}

	return values
}

func Decode(value []byte, into interface{}) error {
	return json.Unmarshal(value, into)
}

func Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}