package oplog

import (
	"github.com/lonelycode/yzma/db"
	"github.com/lonelycode/yzma/types/crdt"
	"github.com/satori/go.uuid"
	"os"
	"testing"
	"time"
)

func NewDB() (*db.DB, string) {
	fName := uuid.NewV4().String()
	d, err := db.New(fName)
	if err != nil {
		panic(err)
	}

	d.Options.CollisionStrategy = crdt.LWWStrat

	return d, fName
}

func teardown(d *db.DB, fName string) {
	d.Close()

	if _, err := os.Stat(fName); err == nil {
		err := os.Remove(fName)
		if err != nil {
			panic(err)
		}
	}
}

func TestWrites(t *testing.T) {
	d1, fn1 := NewDB()
	defer os.Remove(fn1)

	d2, fn2 := NewDB()
	defer os.Remove(fn2)

	h1 := Handler{}
	h1Input := make(chan *OpLog)
	h1.SetReplicaChannel(h1Input)

	h2 := Handler{}
	h2Input := make(chan *OpLog)
	h2.SetReplicaChannel(h2Input)

	rep1 := &InAppReplicator{
		Buffer: h2Input,
	}

	rep2 := &InAppReplicator{
		Buffer: h1Input,
	}

	h1.rep = rep1
	h2.rep = rep2

	h1.Start(d1)
	h2.Start(d2)

	h1.Add("k1", "foo")
	h1.Add("k1", "bar")
	h1.Add("k1", "baz")
	h2.Add("k1", "baz2")

	time.Sleep(100 * time.Millisecond)
	h2.Remove("k1")

	time.Sleep(100 * time.Millisecond)
	x, ok := d1.Load("k1")

	log.Info(x.Extract())
	log.Info(ok)

	//time.Sleep(100 * time.Millisecond)
	//d.Db.View(func(tx *bolt.Tx) error {
	//	c := tx.Bucket([]byte(db.KEYS)).Cursor()
	//	addPrefix := []byte("")
	//	for k, _ := c.Seek(addPrefix); k != nil && bytes.HasPrefix(k, addPrefix); k, _ = c.Next() {
	//		log.Warn(string(k))
	//	}
	//
	//	return nil
	//})
}
