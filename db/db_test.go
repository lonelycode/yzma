package db

import (
	"fmt"
	"github.com/lonelycode/yzma/types/crdt"
	"github.com/satori/go.uuid"
	"os"
	"testing"
)

func NewORSet() (*DB, string) {
	fName := uuid.NewV4().String()
	d, err := New(fName)
	if err != nil {
		panic(err)
	}

	d.Options.CollisionStrategy = crdt.LWWStrat

	return d, fName
}

func teardown(d *DB, fName string) {
	d.Close()

	if _, err := os.Stat(fName); err == nil {
		err := os.Remove(fName)
		if err != nil {
			panic(err)
		}
	}
}

func TestMultiAdd(t *testing.T) {
	orSet, n := NewORSet()
	defer teardown(orSet, n)

	orSet2, n2 := NewORSet()
	defer teardown(orSet2, n2)

	orSet.Add("k1", "foo")
	orSet.Add("k1", "bar")
	orSet.Add("k1", "baz")

	dat, ok := orSet.Load("k1")
	fmt.Println(ok)
	fmt.Println(dat.Extract())
}

func TestORSetAddContains(t *testing.T) {
	orSet, n := NewORSet()
	defer teardown(orSet, n)

	var testValue = "object"

	_, ok := orSet.Load(testValue)
	if ok {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}

	orSet.Add(testValue, "foo")

	_, ok = orSet.Load(testValue)
	if !ok {
		t.Errorf("Expected set to contain: %v, but not found", testValue)
	}
}

func TestORSetAddRemoveContains(t *testing.T) {
	orSet, n := NewORSet()
	defer teardown(orSet, n)


	var testValue string = "object"
	orSet.Add(testValue, "foo")

	orSet.Remove(testValue)

	_, ok := orSet.Load(testValue)
	if ok {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}
}

func TestORSetAddRemoveAddContains(t *testing.T) {
	orSet, n := NewORSet()
	defer teardown(orSet, n)



	var testValue string = "object"

	orSet.Add(testValue, "foo")
	orSet.Remove(testValue)
	orSet.Add(testValue, "foo")

	v, ok := orSet.Load(testValue)
	if !ok {
		t.Errorf("Expected set to contain: %v, but not found (%v)", testValue, v.Extract())
	}
}

func TestORSetAddAddRemoveContains(t *testing.T) {
	orSet, n := NewORSet()
	defer teardown(orSet, n)

	var testValue string = "object"

	orSet.Add(testValue, "foo")
	orSet.Add(testValue, "foo")
	orSet.Remove(testValue)

	_, ok := orSet.Load(testValue)
	if ok {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}
}

//func TestReplication(t *testing.T) {
//	db1, n1 := NewORSet()
//	db2, n2 := NewORSet()
//	defer teardown(db1, n1)
//	defer teardown(db2, n2)
//
//	var jbufDb1 = rbolt.NewJournalBuffer(db1.Db)
//	var jbufDb2 = rbolt.NewJournalBuffer(db2.Db)
//	var transportDb1 = &rbolt.LocalTransport{JournalBuffer: jbufDb1}
//	var transportDb2 = &rbolt.LocalTransport{JournalBuffer: jbufDb2}
//
//	db1.Replication.Transport = transportDb1
//	db2.Replication.Transport = transportDb2
//
//	db1.Add("foo", "bar")
//	db2.Add("foo", "baz")
//	db1.Add("foo2", "bazington")
//
//	if err := jbufDb1.Flush(); err != nil {
//		t.Fatal(err)
//	}
//
//	if err := jbufDb2.Flush(); err != nil {
//		t.Fatal(err)
//	}
//
//	v, ok := db1.Load("foo")
//	log.Warn(ok)
//	log.Info(v.Extract())
//
//	v2, ok2 := db2.Load("foo")
//	log.Warn(ok2)
//	log.Info(v2.Extract())
//
//	r1, ok3 := db2.Load("foo2")
//	log.Warn(ok3)
//	log.Info(r1.Extract())
//
//	db1.Db.View(func(tx *bolt.Tx) error {
//		c := tx.Bucket([]byte(KEYS)).Cursor()
//		addPrefix := []byte("")
//		for k, _ := c.Seek(addPrefix); k != nil && bytes.HasPrefix(k, addPrefix); k, _ = c.Next() {
//			log.Warn(string(k))
//		}
//
//		return nil
//	})
//
//	db2.Db.View(func(tx *bolt.Tx) error {
//		c := tx.Bucket([]byte(KEYS)).Cursor()
//		addPrefix := []byte("")
//		for k, _ := c.Seek(addPrefix); k != nil && bytes.HasPrefix(k, addPrefix); k, _ = c.Next() {
//			log.Warn(string(k))
//		}
//
//		return nil
//	})
//
//}