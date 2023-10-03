package hostsfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	lo := newLookup()
	lo.add("test", 1)
	assert.Len(t, lo.get("test"), 1)
	lo.remove("test", 1)
	assert.Len(t, lo.get("test"), 0)
}
