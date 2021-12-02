package hostsfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemInSlice(t *testing.T) {
	liststring := []string{"hello", "brah"}
	assert.True(t, itemInSliceString("hello", liststring))
	assert.False(t, itemInSliceString("goodbye", liststring))

	intlist := []int{1, 2}
	assert.True(t, itemInSliceInt(2, intlist))
	assert.False(t, itemInSliceInt(3, intlist))
}

func TestRemoveFromSlice(t *testing.T) {
	stringlist := []string{"why", "hello", "there"}
	removeFromSliceString("why", stringlist)
	assert.False(t, itemInSliceString("why", stringlist))
	assert.True(t, itemInSliceString("hello", stringlist))
	assert.True(t, itemInSliceString("there", stringlist))

	intlist := []int{1, 2, 3}
	removeFromSliceInt(2, intlist)
	assert.False(t, itemInSliceInt(2, intlist))
	assert.True(t, itemInSliceInt(1, intlist))
	assert.True(t, itemInSliceInt(3, intlist))
}
