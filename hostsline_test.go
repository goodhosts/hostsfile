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

func TestHostsline_CommentWithMultipleHashes(t *testing.T) {
	// Test that comments with multiple # characters are preserved correctly
	raw := "127.0.0.1 localhost # comment with # symbol in it"
	hl := NewHostsLine(raw)

	assert.Equal(t, "127.0.0.1", hl.IP)
	assert.Equal(t, []string{"localhost"}, hl.Hosts)
	assert.Equal(t, " comment with # symbol in it", hl.Comment)
	assert.Equal(t, raw, hl.ToRaw())

	// Test another case
	raw2 := "192.168.1.1 host1 host2 # first # second # third"
	hl2 := NewHostsLine(raw2)

	assert.Equal(t, "192.168.1.1", hl2.IP)
	assert.Equal(t, []string{"host1", "host2"}, hl2.Hosts)
	assert.Equal(t, " first # second # third", hl2.Comment)
	assert.Equal(t, raw2, hl2.ToRaw())
}
