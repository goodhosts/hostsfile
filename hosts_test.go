package hostsfile

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
)

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func newHosts() *Hosts {
	return &Hosts{
		ips:   lookup{l: make(map[string][]int)},
		hosts: lookup{l: make(map[string][]int)},
	}
}

func newMacOSXDefault() *Hosts {
	h := newHosts()
	if err := h.loadString(`##
# Host Database
#
# localhost is used to configure the loopback interface
# when the system is booting.  Do not change this entry.
##
127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost`); err != nil {
		return newHosts()
	}

	return h
}

func newWindowsDefault() *Hosts {
	h := newHosts()
	if err := h.loadString(`# Copyright (c) 1993-2009 Microsoft Corp.
#
# This is a sample HOSTS file used by Microsoft TCP/IP for Windows.
#
# This file contains the mappings of IP addresses to host names. Each
# entry should be kept on an individual line. The IP address should
# be placed in the first column followed by the corresponding host name.
# The IP address and the host name should be separated by at least one
# space.
#
# Additionally, comments (such as these) may be inserted on individual
# lines or following the machine name denoted by a '#' symbol.
#
# For example:
#
# localhost name resolution is handled within DNS itself.
# 102.54.94.97 rhino.acme.com # source server
# 38.25.63.10 x.acme.com # x client host
# 127.0.0.1 localhost
# ::1 localhost`); err != nil {
		return newHosts()
	}

	return h
}

func newProxmoxDefault() *Hosts {
	h := newHosts()
	if err := h.loadString(`[::1 ip6-localhost ip6-loopback]
fe00::0 
ff00::0 
ff02::1 
ff02::2 
ff02::3`); err != nil {
		return newHosts()
	}

	return h
}

func newMAMPDefault() *Hosts {
	h := newHosts()
	if err := h.loadString(`127.0.0.1	scratch.test	# MAMP PRO - Do NOT remove this entry!
::1		scratch.test	# MAMP PRO - Do NOT remove this entry!
127.0.0.1	clean.test	# MAMP PRO - Do NOT remove this entry!
::1		clean.test	# MAMP PRO - Do NOT remove this entry!
127.0.0.1	cnmd.test	# MAMP PRO - Do NOT remove this entry!
::1		cnmd.test	# MAMP PRO - Do NOT remove this entry!
127.0.0.1	boilerplate.test	# MAMP PRO - Do NOT remove this entry!
::1		boilerplate.test	# MAMP PRO - Do NOT remove this entry!
127.0.0.1	macster.local	# MAMP PRO - Do NOT remove this entry!
::1		macster.local	# MAMP PRO - Do NOT remove this entry!`); err != nil {
		return newHosts()
	}

	return h
}

func Test_DefaultHosts(t *testing.T) {
	mac := newMacOSXDefault()
	assert.Len(t, mac.Lines, 9)

	win := newWindowsDefault()
	assert.Len(t, win.Lines, 20)

	pve := newProxmoxDefault()
	assert.Len(t, pve.Lines, 6)

	mamp := newMAMPDefault()
	assert.Len(t, mamp.Lines, 10)
}

func Test_NewHosts(t *testing.T) {
	hosts, err := NewHosts()
	assert.NoError(t, err)
	assert.NotEqual(t, "", hosts.Path)

	// test env var
	expected := os.ExpandEnv(filepath.FromSlash("./test"))
	f, err := os.Create(expected)
	assert.Nil(t, err)
	defer func() {
		if err := f.Close(); err != nil {
			log.Panic(err)
		}
		if err := os.Remove(expected); err != nil {
			log.Panic(err)
		}
	}()

	assert.Nil(t, os.Setenv("HOSTS_PATH", expected))
	hosts, err = NewHosts()
	assert.NoError(t, err)
	assert.Equal(t, expected, hosts.Path)

	// test is writeable
	assert.True(t, hosts.IsWritable())
	hosts.Path = "./noexist"
	assert.False(t, hosts.IsWritable())

	// test bad load
	assert.Error(t, hosts.Load())
}

func Test_NewCustomHosts(t *testing.T) {
	// bad file
	_, err := NewCustomHosts("./noexist")
	assert.Error(t, err)
}

func TestHostsLine_IsComment(t *testing.T) {
	comment := "   # This is a comment   "
	line := NewHostsLine(comment)
	assert.True(t, line.IsComment())
}

func TestNewHostsLine(t *testing.T) {

	var hlTests = []struct {
		input   string
		output  HostsLine
		asserts func(t *testing.T, hl HostsLine)
	}{
		{
			input: "",
			output: HostsLine{
				Raw:     "",
				Comment: "",
				Err:     nil,
			},
			asserts: func(t *testing.T, hl HostsLine) {
				assert.False(t, hl.IsComment())
				assert.False(t, hl.IsValid())
			},
		}, {
			input: "   # This is a comment   ",
			output: HostsLine{
				Raw:     "   # This is a comment   ",
				Comment: " This is a comment   ",
				Err:     nil,
			},
			asserts: func(t *testing.T, hl HostsLine) {
				assert.True(t, hl.HasComment())
				assert.False(t, hl.IsValid())
			},
		}, {
			input: "127.0.0.1 test1 test2   # This is a comment   ",
			output: HostsLine{
				Raw: "127.0.0.1 test1 test2   # This is a comment   ",
				IP:  "127.0.0.1",
				Hosts: []string{
					"test1", "test2",
				},
				Comment: " This is a comment   ",
				Err:     nil,
			},
			asserts: func(t *testing.T, hl HostsLine) {
				assert.True(t, hl.HasComment())
				assert.False(t, hl.IsMalformed())
			},
		}, {
			// bad ip parse
			input: "127.x.x.1 test1 test2   # This is a comment   ",
			output: HostsLine{
				Raw: "127.x.x.1 test1 test2   # This is a comment   ",
				IP:  "127.x.x.1",
				Hosts: []string{
					"test1", "test2",
				},
				Comment: " This is a comment   ",
				Err:     errors.New("bad hosts line: \"127.x.x.1 test1 test2   \""),
			},
			asserts: func(t *testing.T, hl HostsLine) {
				assert.True(t, hl.IsValid()) // technically valid, just had an ip parse error... ?
				assert.True(t, hl.IsMalformed())
			},
		},
	}

	for _, tt := range hlTests {
		hl := NewHostsLine(tt.input)
		assert.Equal(t, tt.output, hl)
		if nil != tt.asserts {
			tt.asserts(t, hl)
		}
	}
}

func TestHosts_Has(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.AddRaw("127.0.0.1 yadda", "10.0.0.7 nada"))
	assert.True(t, hosts.Has("10.0.0.7", "nada"))
	assert.False(t, hosts.Has("10.0.0.7", "shuda"))
}

func TestHosts_Remove(t *testing.T) {
	// when last host ip combo
	expectedLines := []HostsLine{NewHostsLine("127.0.0.1 yadda")}

	hosts := newHosts()
	assert.Nil(t, hosts.AddRaw("127.0.0.1 yadda", "10.0.0.7 nada"))
	assert.Nil(t, hosts.Remove("10.0.0.7", "nada"))
	assert.Equal(t, expectedLines, hosts.Lines)

	// when ip has other hosts
	expectedLines = []HostsLine{NewHostsLine("127.0.0.1 yadda"), NewHostsLine("10.0.0.7 brada")}
	hosts = newHosts()
	assert.Nil(t, hosts.AddRaw("127.0.0.1 yadda", "10.0.0.7 nada brada"))
	assert.Nil(t, hosts.Remove("10.0.0.7", "nada"))
	assert.Equal(t, expectedLines, hosts.Lines)

	// remove multiple entries
	hosts = newHosts()
	assert.Nil(t, hosts.AddRaw("127.0.0.1 yadda nadda prada"))
	assert.Nil(t, hosts.Remove("127.0.0.1", "yadda", "prada"))
	assert.Equal(t, hosts.Lines[0].Raw, "127.0.0.1 nadda")

	// nothing to remove
	assert.Nil(t, hosts.Remove("127.0.0.1"))

	// remove bad ip
	assert.Error(t, hosts.Remove("not an ip"))
}

func TestHosts_HasHostname(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))
	assert.Nil(t, hosts.Add("10.0.0.7", "nada"))
	assert.True(t, hosts.HasHostname("nada"))
	assert.False(t, hosts.HasHostname("shuda"))
}

func TestHosts_RemoveByIp(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))
	assert.Nil(t, hosts.Add("10.0.0.7", "nada"))
	assert.Nil(t, hosts.Add("42.42.42.42", "foo"))
	assert.Nil(t, hosts.Add("10.0.0.255", "bar"))

	// remove nothing
	assert.Nil(t, hosts.RemoveByIp("192.168.1.1"))
	assert.Len(t, hosts.Lines, 4)
	assert.Len(t, hosts.ips.l, 4)
	assert.Len(t, hosts.hosts.l, 4)

	// remove 1
	assert.Nil(t, hosts.RemoveByIp("10.0.0.255"))
	assert.Len(t, hosts.Lines, 3)
	assert.Len(t, hosts.ips.l, 3)
	assert.Len(t, hosts.hosts.l, 3)

	// remove 1
	assert.Nil(t, hosts.RemoveByIp("10.0.0.7"))
	assert.Len(t, hosts.Lines, 2)
	assert.Len(t, hosts.ips.l, 2)
	assert.Len(t, hosts.hosts.l, 2)

	// remove 1
	assert.Nil(t, hosts.RemoveByIp("127.0.0.1"))
	assert.Len(t, hosts.Lines, 1)
	assert.Len(t, hosts.ips.l, 1)
	assert.Len(t, hosts.hosts.l, 1)

	// remove 0
	assert.Nil(t, hosts.RemoveByIp("10.0.0.7"))
	assert.Len(t, hosts.Lines, 1)
	assert.Len(t, hosts.ips.l, 1)
	assert.Len(t, hosts.hosts.l, 1)
}

func TestHosts_RemoveByHostname(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))
	assert.Nil(t, hosts.Add("168.1.1.1", "yadda"))

	assert.Nil(t, hosts.RemoveByHostname("yadda"))
	assert.False(t, hosts.HasHostname("yadda"))

	// remove if hostname doesn't exist
	hosts = newHosts()
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))

	assert.False(t, hosts.HasHostname("prada"))
	assert.Nil(t, hosts.RemoveByHostname("prada"))

	// remove if exists
	assert.True(t, hosts.HasHostname("yadda"))
	assert.Nil(t, hosts.RemoveByHostname("yadda"))
	assert.False(t, hosts.HasHostname("yadda"))

	// remove hostname and clean up the IP address if
	// it was the only name/alias on the line
	hosts = newHosts()
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))
	assert.Nil(t, hosts.Add("168.1.1.1", "prada"))
	assert.Nil(t, hosts.Add("1.2.3.4", "foo", "bar"))

	assert.Nil(t, hosts.RemoveByHostname("yadda"))
	assert.Len(t, hosts.Lines, 2)
	assert.True(t, hosts.HasHostname("prada"))
	assert.True(t, hosts.HasHostname("foo"))
	assert.True(t, hosts.HasHostname("bar"))
	assert.Equal(t, hosts.hosts.l["prada"], []int{0})
	assert.Equal(t, hosts.hosts.l["foo"], []int{1})
	assert.Equal(t, hosts.hosts.l["bar"], []int{1})

	assert.Nil(t, hosts.RemoveByHostname("foo"))
	assert.Len(t, hosts.Lines, 2)
	assert.True(t, hosts.HasHostname("prada"))
	assert.True(t, hosts.HasHostname("bar"))

	assert.Nil(t, hosts.RemoveByHostname("bar"))
	assert.Len(t, hosts.Lines, 1)
	assert.True(t, hosts.HasHostname("prada"))
	assert.Equal(t, hosts.hosts.l["prada"], []int{0})
}

func TestHosts_HasIp(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))
	assert.Nil(t, hosts.Add("168.1.1.1", "yadda"))

	// add should have removed yadda from 127.0.0.1
	assert.False(t, hosts.HasIp("127.0.0.1"))
	assert.Len(t, hosts.ips.l, 1)
	assert.Len(t, hosts.hosts.l, 1)
	assert.True(t, hosts.HasIp("168.1.1.1"))
}

func TestHosts_LineWithTrailingComment(t *testing.T) {
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
		hosts := newHosts()
		assert.Nil(t, hosts.AddRaw(test.given))
		assert.Nil(t, hosts.Add(test.addIp, test.addHost))
		assert.Equal(t, hosts.Lines[0].Raw, test.expected)
	}
}

func TestHosts_LineWithComments(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.AddRaw("#This is the first comment",
		"127.0.0.1 prada",
		"#This is the second comment",
		"127.0.0.2 tada #HostLine with trailing comment",
		"#This is third comment"))

	for _, hostLine := range hosts.Lines {
		assert.Equal(t, hostLine.ToRaw(), hostLine.Raw)
	}
}

func TestHosts_Clean(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.AddRaw("127.0.0.2 prada yadda #comment1", "127.0.0.2 tada abba #comment2"))
	hosts.Clean()
	assert.Equal(t, len(hosts.Lines), 1)
	assert.Equal(t, hosts.Lines[0].Comment, "comment1 comment2")
	assert.Equal(t, hosts.Lines[0].ToRaw(), "127.0.0.2 abba prada tada yadda #comment1 comment2")
}

func TestHosts_Add(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.Add("127.0.0.2", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))  // valid use with variatic args
	assert.Error(t, assert.AnError, hosts.Add("127.0.0.2", "host11 host12 host13 host14 host15 host16 host17 host18 hosts19 hosts20")) // invalid use
	assert.Len(t, hosts.Lines, 1)
	assert.Nil(t, hosts.Add("127.0.0.3", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))
	assert.Len(t, hosts.Lines, 1)
	assert.Error(t, assert.AnError, hosts.Add("127.0.0.3", "invalid hostname"))
	assert.Error(t, assert.AnError, hosts.Add("127.0.0.3", ".invalid*hostname"))

	// don't add if the combo ip/host exists somewhere in the file
	hosts = newHosts()
	assert.Nil(t, hosts.AddRaw("127.0.0.1 tom.test", "127.0.0.1 tom.test example.test"))
	assert.Nil(t, hosts.Add("127.0.0.1", "example.test"))
	assert.Equal(t, hosts.Lines[0].Raw, "127.0.0.1 tom.test")

	// call with no hosts
	hosts = newHosts()
	assert.Nil(t, hosts.AddRaw("127.0.0.1 yadda", "10.0.0.7 nada"))
	hosts.Lines[1] = HostsLine{
		IP:    "not an ip",
		Hosts: []string{"nada"},
	}
	assert.Error(t, hosts.Add("192.168.1.1", "nada"))

	// reset and try adding the same hosts multiple times to two ips
	hosts = newHosts()

	// add 10 hosts
	assert.Nil(t, hosts.Add("127.0.0.2", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))
	assert.Len(t, hosts.Lines, 1)
	assert.Len(t, hosts.hosts.l, 10)
	assert.Len(t, hosts.ips.l, 1)

	// add the same thing twice, should be no additions.
	assert.Nil(t, hosts.Add("127.0.0.2", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))
	assert.Len(t, hosts.Lines, 1)
	assert.Len(t, hosts.hosts.l, 10)
	assert.Len(t, hosts.ips.l, 1)

	// add a new ip with 10 hosts, should remove first ip
	assert.Nil(t, hosts.Add("127.0.0.3", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))
	assert.False(t, hosts.HasIp("127.0.0.2"))
	assert.Len(t, hosts.Lines, 1)
	assert.Len(t, hosts.hosts.l, 10)
	assert.Len(t, hosts.ips.l, 1)

	// make sure adding a duplicate host removes it form the previous ip
	expectedLines := []HostsLine{NewHostsLine("10.0.0.7 nada yadda brada")}
	hosts = newHosts()
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))
	assert.Nil(t, hosts.Add("10.0.0.7", "nada", "yadda"))
	assert.Nil(t, hosts.Add("10.0.0.7", "brada", "yadda"))
	assert.Len(t, hosts.ips.l, 1)
	assert.Len(t, hosts.hosts.l, 3)
	assert.Equal(t, expectedLines, hosts.Lines)
}

func TestHosts_HostsPerLine(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.Add("127.0.0.2", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))
	assert.Nil(t, hosts.Add("127.0.0.2", "host11", "host12", "host13", "host14", "host15", "host16", "host17", "host18", "host19", "hosts20"))
	hosts.HostsPerLine(1)
	assert.Len(t, hosts.Lines, 20)
	hosts.HostsPerLine(2)
	assert.Len(t, hosts.Lines, 10)
	hosts.HostsPerLine(9) // windows
	assert.Len(t, hosts.Lines, 3)
	hosts.HostsPerLine(50) // all in one
	assert.Len(t, hosts.Lines, 1)

	hosts = newHosts()
	assert.Nil(t, hosts.Add("127.0.0.2", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))
	hosts.HostsPerLine(8)
	assert.Nil(t, hosts.Add("127.0.0.2", "host1", "host2", "host3", "host4", "host5", "host6", "host7", "host8", "host9", "hosts10"))

}

func BenchmarkHosts_Add10k(b *testing.B) {
	benchmarkHosts_Add(10000, b)
}

func BenchmarkHosts_Add25k(b *testing.B) {
	benchmarkHosts_Add(25000, b)
}

func BenchmarkHosts_Add50k(b *testing.B) {
	benchmarkHosts_Add(50000, b)
}

func BenchmarkHosts_Add250k(b *testing.B) {
	benchmarkHosts_Add(250000, b)
}

func benchmarkHosts_Add(c int, b *testing.B) {
	hosts, err := NewCustomHosts("hostsfile")
	assert.Nil(b, err)
	for i := 0; i < c; i++ {
		assert.Nil(b, hosts.Add(fake.IPv4(), randomString(63)))
	}
}

func BenchmarkHosts_Flush50k(b *testing.B) {
	benchmarkHosts_Flush(5, b)
}

func BenchmarkHosts_Flush100k(b *testing.B) {
	benchmarkHosts_Flush(10, b)
}

func BenchmarkHosts_Flush250k(b *testing.B) {
	benchmarkHosts_Flush(25, b)
}

func BenchmarkHosts_Flush500k(b *testing.B) {
	benchmarkHosts_Flush(50, b)
}

// benchmarks flushing a hostsfile and confirms the hashmap lookup for ips/hosts is thread save via mutex + locking
func benchmarkHosts_Flush(c int, b *testing.B) {
	_, err := os.Create("hostsfile")
	assert.Nil(b, err)
	hosts, err := NewCustomHosts("hostsfile")
	assert.Nil(b, err)

	wg := sync.WaitGroup{}
	wg.Add(c)
	for i := 0; i < c; i++ {
		go func() {
			for i := 0; i < 10000; i++ {
				assert.Nil(b, hosts.Add(fake.IPv4(), randomString(63)))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	assert.Nil(b, hosts.Flush())
	assert.Nil(b, os.Remove("hostsfile"))
}

func TestHosts_Flush(t *testing.T) {
	f, err := os.Create("hostsfile")
	defer func() {
		assert.Nil(t, f.Close())
		assert.Nil(t, os.Remove("hostsfile"))
	}()

	assert.Nil(t, err)
	hosts, err := NewCustomHosts("./hostsfile")
	assert.Nil(t, err)
	assert.Nil(t, hosts.Add("127.0.0.2", "host1"))
	assert.Equal(t, 1, len(hosts.Lines))
	assert.Equal(t, "127.0.0.2 host1", hosts.Lines[0].Raw)
	assert.Nil(t, hosts.Flush())
	assert.Equal(t, 1, len(hosts.Lines))
	assert.Equal(t, "127.0.0.2 host1", hosts.Lines[0].Raw)

	// bad path can't write
	hosts.Path = ""
	assert.Error(t, hosts.Flush())
}

func TestHosts_Clear(t *testing.T) {
	hosts := newHosts()
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))
	assert.True(t, hosts.HasIp("127.0.0.1"))
	assert.Len(t, hosts.Lines, 1)
	hosts.Clear()
	assert.Len(t, hosts.Lines, 0)
	assert.Nil(t, hosts.Add("127.0.0.1", "yadda"))
	assert.True(t, hosts.HasIp("127.0.0.1"))
	assert.Len(t, hosts.Lines, 1)
}

func TestHosts_RemoveDuplicateHosts(t *testing.T) {
	h := newHosts()
	assert.Nil(t, h.loadString(`127.0.0.1 test1 test1 test2 test2`))
	assert.Len(t, h.Lines, 1)
	assert.Len(t, h.ips.l, 1)
	assert.Len(t, h.hosts.l, 2)

	h.RemoveDuplicateHosts()
	assert.Len(t, h.Lines, 1)
	assert.Len(t, h.ips.l, 1)
	assert.Len(t, h.hosts.l, 2)
	assert.Equal(t, "127.0.0.1 test1 test2"+eol, h.String())

	h = newHosts()
	assert.Nil(t, h.loadString(`127.0.0.1 test1 test1 test2 test2`+eol+`127.0.0.2 test1 test1 test2 test2`+eol))
	assert.Len(t, h.Lines, 2)
	assert.Len(t, h.ips.l, 2)
	assert.Len(t, h.hosts.l, 2)
	assert.Len(t, h.hosts.l["test1"], 4)
	assert.Len(t, h.hosts.l["test2"], 4)

	h.RemoveDuplicateHosts()
	assert.Len(t, h.Lines, 2)
	assert.Len(t, h.ips.l, 2)
	assert.Len(t, h.hosts.l, 2)
	assert.Len(t, h.hosts.l["test1"], 2)
	assert.Len(t, h.hosts.l["test2"], 2)

	assert.Equal(t, "127.0.0.1 test1 test2"+eol+"127.0.0.2 test1 test2"+eol, h.String())
}
