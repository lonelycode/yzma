package crdt

import (
	"fmt"
	"testing"
)


func TestORSet_Update(t *testing.T) {
	set := NewORSet()
	set.Options.CollisionStrategy = LWWStrat

	set2 := NewORSet()
	set2.Options.CollisionStrategy = LWWStrat

	set.Add("k1", "foo")
	set.Add("k2", "bar")
	set.Add("k3", "baz")

	set.AddOrUpdate("k1", "oof")
	set.AddOrUpdate("k2", "rab")
	set.AddOrUpdate("k3", "zab")


	v1, _ := set.Load("k1")
	v2, _ := set.Load("k2")
	v3, _ := set.Load("k3")

	fmt.Println(v1.Extract())
	fmt.Println(v2.Extract())
	fmt.Println(v3.Extract())

	set2.AddOrUpdate("k1", "ofo")
	set2.AddOrUpdate("k1", "ofofo")

	set.AddOrUpdate("k1", "boooong")
	set.Merge(set2)

	v4, _ := set.Load("k1")
	fmt.Println(v4.Extract())
}

func TestORSetAddContains(t *testing.T) {
	orSet := NewORSet()

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
	orSet := NewORSet()

	var testValue string = "object"
	orSet.Add(testValue, "foo")

	orSet.Remove(testValue)

	_, ok := orSet.Load(testValue)
	if ok {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}
}

func TestORSetAddRemoveAddContains(t *testing.T) {
	orSet := NewORSet()

	var testValue string = "object"

	orSet.Add(testValue, "foo")
	orSet.Remove(testValue)
	orSet.Add(testValue, "foo")

	_, ok := orSet.Load(testValue)
	if !ok {
		t.Errorf("Expected set to contain: %v, but not found", testValue)
	}
}

func TestORSetAddAddRemoveContains(t *testing.T) {
	orSet := NewORSet()

	var testValue string = "object"

	orSet.Add(testValue, "foo")
	orSet.Add(testValue, "foo")
	orSet.Remove(testValue)

	_, ok := orSet.Load(testValue)
	if ok {
		t.Errorf("Expected set to not contain: %v, but found", testValue)
	}
}

func TestORSetMerge(t *testing.T) {
	type addRm struct {
		addSet []string
		rmSet  []string
	}

	for _, tt := range []struct {
		setOne  addRm
		setTwo  addRm
		valid   map[string]struct{}
		invalid map[string]struct{}
	}{
		{
			addRm{[]string{"object1"}, []string{}},
			addRm{[]string{}, []string{"object1"}},
			map[string]struct{}{
				"object1": struct{}{},
			},
			map[string]struct{}{},
		},
		{
			addRm{[]string{}, []string{"object1"}},
			addRm{[]string{"object1"}, []string{}},
			map[string]struct{}{
				"object1": struct{}{},
			},
			map[string]struct{}{},
		},
		{
			addRm{[]string{"object1"}, []string{"object1"}},
			addRm{[]string{}, []string{}},
			map[string]struct{}{},
			map[string]struct{}{
				"object1": struct{}{},
			},
		},
		{
			addRm{[]string{}, []string{}},
			addRm{[]string{"object1"}, []string{"object1"}},
			map[string]struct{}{},
			map[string]struct{}{
				"object1": struct{}{},
			},
		},
		{
			addRm{[]string{"object2"}, []string{"object1"}},
			addRm{[]string{"object1"}, []string{"object2"}},
			map[string]struct{}{
				"object1": struct{}{},
				"object2": struct{}{},
			},
			map[string]struct{}{},
		},
		{
			addRm{[]string{"object2", "object1"}, []string{"object1"}},
			addRm{[]string{"object1", "object2"}, []string{"object2"}},
			map[string]struct{}{
				"object1": struct{}{},
				"object2": struct{}{},
			},
			map[string]struct{}{},
		},
		{
			addRm{[]string{"object2", "object1"}, []string{"object1", "object2"}},
			addRm{[]string{"object1", "object2"}, []string{"object2", "object1"}},
			map[string]struct{}{},
			map[string]struct{}{
				"object1": struct{}{},
				"object2": struct{}{},
			},
		},
	} {
		orset1, orset2 := NewORSet(), NewORSet()

		for _, add := range tt.setOne.addSet {
			orset1.Add(add, "foo")
		}

		for _, rm := range tt.setOne.rmSet {
			orset1.Remove(rm)
		}

		for _, add := range tt.setTwo.addSet {
			orset2.Add(add, "foo")
		}

		for _, rm := range tt.setTwo.rmSet {
			orset2.Remove(rm)
		}

		orset1.Merge(orset2)

		for obj, _ := range tt.valid {
			_, ok := orset1.Load(obj)
			if !ok {
				t.Errorf("expected set to contain: %v", obj)
			}
		}

		for obj, _ := range tt.invalid {
			_, ok := orset1.Load(obj)
			if ok {
				t.Errorf("expected set to not contain: %v", obj)
			}
		}
	}
}
