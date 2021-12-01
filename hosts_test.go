package hostsfile

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostsLineIsComment(t *testing.T) {
	comment := "   # This is a comment   "
	line := NewHostsLine(comment)
	result := line.IsComment()
	if !result {
		t.Error(fmt.Sprintf("%s should be a comment", comment))
	}
}

func TestNewHostsLineWithEmptyLine(t *testing.T) {
	line := NewHostsLine("")
	if line.Raw != "" {
		t.Error("Failed to load empty line.")
	}
}

func TestHostsHas(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 nada")}

	// We should find this entry.
	if !hosts.Has("10.0.0.7", "nada") {
		t.Error("Couldn't find entry in hosts file.")
	}

	// We shouldn't find this entry
	if hosts.Has("10.0.0.7", "shuda") {
		t.Error("Found entry that isn't in hosts file.")
	}
}

func TestHostsHasDoesntFindMissingEntry(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 nada")}

	if hosts.Has("10.0.0.7", "brada") {
		t.Error("Found missing entry.")
	}
}

func TestHostsAddWhenIpHasOtherHosts(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 nada yadda")}

	if err := hosts.Add("10.0.0.7", "brada", "yadda"); err != nil {
		t.Error(err)
	}

	expectedLines := []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 nada yadda brada")}

	if !reflect.DeepEqual(hosts.Lines, expectedLines) {
		t.Error("Add entry failed to append entry.")
	}
}

func TestHostsAddWhenIpDoesntExist(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda")}

	if err := hosts.Add("10.0.0.7", "brada", "yadda"); err != nil {
		t.Error(err)
	}

	expectedLines := []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 brada yadda")}

	if !reflect.DeepEqual(hosts.Lines, expectedLines) {
		t.Error("Add entry failed to append entry.")
	}
}

func TestHostsRemoveWhenLastHostIpCombo(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 nada")}

	if err := hosts.Remove("10.0.0.7", "nada"); err != nil {
		t.Error(err)
	}

	expectedLines := []HostsLine{NewHostsLine("127.0.0.1 yadda")}

	if !reflect.DeepEqual(hosts.Lines, expectedLines) {
		t.Error("Remove entry failed to remove entry.")
	}
}

func TestHostsRemoveWhenIpHasOtherHosts(t *testing.T) {
	hosts := new(Hosts)

	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 nada brada")}

	if err := hosts.Remove("10.0.0.7", "nada"); err != nil {
		t.Error(err)
	}

	expectedLines := []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 brada")}

	if !reflect.DeepEqual(hosts.Lines, expectedLines) {
		t.Error("Remove entry failed to remove entry.")
	}
}

func TestHostsRemoveMultipleEntries(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda nadda prada")}

	if err := hosts.Remove("127.0.0.1", "yadda", "prada"); err != nil {
		t.Error(err)
	}
	if hosts.Lines[0].Raw != "127.0.0.1 nadda" {
		t.Error("Failed to remove multiple entries.")
	}
}

func TestHostnamesHas(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 nada")}

	// We should find this entry.
	if !hosts.HasHostname("nada") {
		t.Error("Couldn't find entry in hosts file.")
	}

	// We shouldn't find this entry
	if hosts.HasHostname("shuda") {
		t.Error("Found entry that isn't in hosts file.")
	}
}

func TestHostsRemoveByHostname(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 yadda"),
		NewHostsLine("168.1.1.1 yadda"),
	}

	if err := hosts.RemoveByHostname("yadda"); err != nil {
		t.Error(err)
	}
	// We shouldn't find this entry
	if hosts.HasHostname("yadda") {
		t.Error("Found entry that isn't in hosts file.")
	}
}

func TestHostsRemoveByHostnameWhenHostnameNotExist(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.1 prada"),
	}

	if err := hosts.RemoveByHostname("yadda"); err != nil {
		t.Error(err)
	}

	// We shouldn't find this entry
	if hosts.HasHostname("yadda") {
		t.Error("Found entry that isn't in hosts file.")
	}

	if !hosts.HasHostname("prada") {
		t.Error("Did not find entry that is in hosts file.")
	}
}

func TestHostsLineWithTrailingComment(t *testing.T) {
	tests := []struct {
		given    string
		addIp    string
		addHost  string
		expected string
	}{
		{
			given:    "127.0.0.1 prada #comment",
			addIp:    "127.0.0.1",
			addHost:  "yadda",
			expected: "127.0.0.1 prada yadda #comment",
		},
		{
			given:    "127.0.0.1 prada # comment",
			addIp:    "127.0.0.1",
			addHost:  "yadda",
			expected: "127.0.0.1 prada yadda # comment",
		},
	}

	for _, test := range tests {
		hosts := new(Hosts)
		hosts.Lines = []HostsLine{
			NewHostsLine(test.given),
		}

		if err := hosts.Add(test.addIp, test.addHost); err != nil {
			t.Error(err)
		}

		if hosts.Lines[0].Raw != test.expected {
			t.Errorf("Failed to add new host to line with comment: expected '%s' received '%s'",
				test.expected, hosts.Lines[0].Raw)
		}
	}
}

func TestHostsLineWithComments(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("#This is the first comment"),
		NewHostsLine("127.0.0.1 prada"),
		NewHostsLine("#This is the second comment"),
		NewHostsLine("127.0.0.2 tada #HostLine with trailing comment"),
		NewHostsLine("#This is third comment"),
	}
	for _, hostLine := range hosts.Lines {
		if hostLine.ToRaw() != hostLine.Raw {
			t.Errorf("Conversion to Raw String Failed")
		}
	}
}

func TestHostsClean(t *testing.T) {
	hosts := new(Hosts)
	hosts.Lines = []HostsLine{
		NewHostsLine("127.0.0.2 prada yadda #comment1"),
		NewHostsLine("127.0.0.2 tada abba #comment2"),
	}

	hosts.Clean()
	if len(hosts.Lines) != 1 {
		t.Errorf("Clean failed to combine IPs")
	}

	if hosts.Lines[0].Comment != "comment1 comment2" {
		t.Errorf("Clean did not update Comment properly: %s", hosts.Lines[0].Comment)
	}

	if hosts.Lines[0].ToRaw() != "127.0.0.2 abba prada tada yadda #comment1 comment2" {
		t.Errorf("Clean did not update Raw properly: %s", hosts.Lines[0].ToRaw())
	}
}

func TestHosts_Add(t *testing.T) {
	hosts := new(Hosts)
	assert.Nil(t, hosts.Add("127.0.0.2", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))  // valid use with variatic args
	assert.Error(t, assert.AnError, hosts.Add("127.0.0.2", "host11 host12 host13 host14 host15 host16 host17 host18 hosts19 hosts20")) // invalid use
	assert.Len(t, hosts.Lines, 1)
	assert.Nil(t, hosts.Add("127.0.0.3", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))
	assert.Len(t, hosts.Lines, 2)
	assert.Error(t, assert.AnError, hosts.Add("127.0.0.3", "invalid hostname"))
	assert.Error(t, assert.AnError, hosts.Add("127.0.0.3", ".invalid*hostname"))
}

func TestHosts_HostsPerLine(t *testing.T) {
	hosts := new(Hosts)
	assert.Nil(t, hosts.Add("127.0.0.2", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))
	assert.Nil(t, hosts.Add("127.0.0.2", "host11", "host12", "host13", "host14", "host15", "host16", "host17", "host18", "host19", "hosts20"))
	hosts.HostsPerLine(1)
	assert.Len(t, hosts.Lines, 20)
	hosts.HostsPerLine(2)
	assert.Len(t, hosts.Lines, 10)
	hosts.HostsPerLine(9) // windows
	assert.Len(t, hosts.Lines, 3)
}
