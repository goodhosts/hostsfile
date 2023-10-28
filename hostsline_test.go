package hostsfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostsline_String(t *testing.T) {
	hl := HostsLine{IP: "127.0.0.1", Hosts: []string{"localhost"}}
	assert.Equal(t, "127.0.0.1 localhost", hl.String())
}

func TestHosts_combine(t *testing.T) {
	hl1 := HostsLine{IP: "127.0.0.1", Hosts: []string{"test1"}}
	hl2 := HostsLine{IP: "127.0.0.1", Hosts: []string{"test2"}}
	assert.Len(t, hl1.Hosts, 1)

	hl1.combine(hl2)
	assert.Len(t, hl1.Hosts, 2)
	assert.Equal(t, "127.0.0.1 test1 test2", hl1.String())

	// change to combine when deprecated
	hl2.Combine(hl1) // should have dupes removed
	assert.Equal(t, "127.0.0.1 test2 test1 test2", hl2.String())
}
