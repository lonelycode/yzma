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

//func TestORSetMerge(t *testing.T) {
//	type addRm struct {
//		addSet []string
//		rmSet  []string
//	}
//
//	for _, tt := range []struct {
//		setOne  addRm
//		setTwo  addRm
//		valid   map[string]struct{}
//		invalid map[string]struct{}
//	}{
//		{
//			addRm{[]string{"object1"}, []string{}},
//			addRm{[]string{}, []string{"object1"}},
//			map[string]struct{}{
//				"object1": struct{}{},
//			},
//			map[string]struct{}{},
//		},
//		{
//			addRm{[]string{}, []string{"object1"}},
//			addRm{[]string{"object1"}, []string{}},
//			map[string]struct{}{
//				"object1": struct{}{},
//			},
//			map[string]struct{}{},
//		},
//		{
//			addRm{[]string{"object1"}, []string{"object1"}},
//			addRm{[]string{}, []string{}},
//			map[string]struct{}{},
//			map[string]struct{}{
//				"object1": struct{}{},
//			},
//		},
//		{
//			addRm{[]string{}, []string{}},
//			addRm{[]string{"object1"}, []string{"object1"}},
//			map[string]struct{}{},
//			map[string]struct{}{
//				"object1": struct{}{},
//			},
//		},
//		{
//			addRm{[]string{"object2"}, []string{"object1"}},
//			addRm{[]string{"object1"}, []string{"object2"}},
//			map[string]struct{}{
//				"object1": struct{}{},
//				"object2": struct{}{},
//			},
//			map[string]struct{}{},
//		},
//		{
//			addRm{[]string{"object2", "object1"}, []string{"object1"}},
//			addRm{[]string{"object1", "object2"}, []string{"object2"}},
//			map[string]struct{}{
//				"object1": struct{}{},
//				"object2": struct{}{},
//			},
//			map[string]struct{}{},
//		},
//		{
//			addRm{[]string{"object2", "object1"}, []string{"object1", "object2"}},
//			addRm{[]string{"object1", "object2"}, []string{"object2", "object1"}},
//			map[string]struct{}{},
//			map[string]struct{}{
//				"object1": struct{}{},
//				"object2": struct{}{},
//			},
//		},
//	} {
//		orset1, orset2 := NewORSet(), NewORSet()
//
//		for _, add := range tt.setOne.addSet {
//			orset1.Add(add, "foo")
//		}
//
//		for _, rm := range tt.setOne.rmSet {
//			orset1.Remove(rm)
//		}
//
//		for _, add := range tt.setTwo.addSet {
//			orset2.Add(add, "foo")
//		}
//
//		for _, rm := range tt.setTwo.rmSet {
//			orset2.Remove(rm)
//		}
//
//		orset1.Merge(orset2)
//
//		for obj, _ := range tt.valid {
//			_, ok := orset1.Load(obj)
//			if !ok {
//				t.Errorf("expected set to contain: %v", obj)
//			}
//		}
//
//		for obj, _ := range tt.invalid {
//			_, ok := orset1.Load(obj)
//			if ok {
//				t.Errorf("expected set to not contain: %v", obj)
//			}
//		}
//	}
//}
