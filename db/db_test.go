package db

import (
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
