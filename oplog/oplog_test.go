package oplog

import (
	"github.com/lonelycode/yzma/db"
	"github.com/lonelycode/yzma/types/crdt"
	"github.com/satori/go.uuid"
	"os"
	"testing"
	"time"
)

func NewDB() (*Handler, *db.DB, string) {
	fName := uuid.NewV4().String()
	d, err := db.New(fName)
	if err != nil {
		panic(err)
	}

	d.Options.CollisionStrategy = crdt.LWWStrat
	h := &Handler{}
	h.Start(d)

	return h, d, fName
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

func TestDB_AddContains(t *testing.T) {
	handler, db, n := NewDB()
	defer teardown(db, n)

	var testValue = "object"

	_, ok := db.Load(testValue)
	if ok {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}

	handler.Add(testValue, []byte("foo"), "")

	time.Sleep(100 * time.Millisecond)
	_, ok = db.Load(testValue)
	if !ok {
		t.Errorf("Expected set to contain: %v, but not found", testValue)
	}
}

func TestDB_AddRemoveContains(t *testing.T) {
	handler, db, n := NewDB()
	defer teardown(db, n)

	var testValue string = "object"
	handler.Add(testValue, []byte("foo"), "")

	time.Sleep(100 * time.Millisecond)
	handler.Remove(testValue)
	time.Sleep(100 * time.Millisecond)

	_, ok := db.Load(testValue)
	if ok {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}
}

func TestDB_AddRemoveAddContains(t *testing.T) {
	handler, db, n := NewDB()
	defer teardown(db, n)

	var testValue string = "object"

	handler.Add(testValue, []byte("foo"), "")
	handler.Remove(testValue)
	handler.Add(testValue, []byte("foo"), "")

	v, ok := db.Load(testValue)
	if !ok {
		d, _ := v.Extract()
		t.Errorf("Expected set to contain: %v, but not found (%v)", testValue, d)
	}
}

func TestDB_AddAddRemoveContains(t *testing.T) {
	handler, db, n := NewDB()
	defer teardown(db, n)

	var testValue string = "object"

	handler.Add(testValue, []byte("foo"), "")
	handler.Add(testValue, []byte("foo"), "")

	// TODO: This isn't great, the writes are too fast for the removes to take effect
	time.Sleep(100 * time.Millisecond)
	handler.Remove(testValue)
	time.Sleep(100 * time.Millisecond)

	_, ok := db.Load(testValue)
	if ok {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}
}

