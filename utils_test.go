package hostsfile

import (
	"fmt"
	"testing"
)

func TestItemInSlice(t *testing.T) {
	item := "this"
	list := []string{"hello", "brah"}
	result := itemInSlice("goodbye", list)
	if result {
		t.Error(fmt.Sprintf("'%s' should not have been found in slice.", item))
	}

	item = "hello"
	result = itemInSlice(item, list)
	if !result {
		t.Error(fmt.Sprintf("'%s' should have been found in slice.", item))
	}
}

func TestRemoveFromSlice(t *testing.T) {
	item := "why"
	list := []string{"why", "hello", "there"}
	removeFromSlice("why", list)
	result :=itemInSlice("why", list)
	if result {
		t.Error(fmt.Sprintf("'%s' should not have been found in slice.", item))
	}

	item = "hello"
	result = itemInSlice(item, list)
	if !result {
		t.Error(fmt.Sprintf("'%s' should have been found in slice.", item))
	}

	item = "there"
	result = itemInSlice(item, list)
	if !result {
		t.Error(fmt.Sprintf("'%s' should have been found in slice.", item))
	}
}
