package hostsfile

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/dimchansky/utfbom"
)

type Hosts struct {
	Path  string
	Lines []HostsLine
	ips   lookup
	hosts lookup
}

// NewHosts return a new instance of Hosts using the default hosts file path.
func NewHosts() (*Hosts, error) {
	osHostsFilePath := os.ExpandEnv(filepath.FromSlash(HostsFilePath))

	if env, isset := os.LookupEnv("HOSTS_PATH"); isset && len(env) > 0 {
		osHostsFilePath = os.ExpandEnv(filepath.FromSlash(env))
	}

	return NewCustomHosts(osHostsFilePath)
}

// NewCustomHosts return a new instance of Hosts using a custom hosts file path.
func NewCustomHosts(osHostsFilePath string) (*Hosts, error) {
	hosts := &Hosts{
		Path:  osHostsFilePath,
		ips:   newLookup(),
		hosts: newLookup(),
	}

	if err := hosts.Load(); err != nil {
		return hosts, err
	}

	return hosts, nil
}

// String get a string of the contents of the contents to put in the hosts file
func (h *Hosts) String() string {
	buf := new(bytes.Buffer)
	for _, line := range h.Lines {
		if _, err := fmt.Fprintf(buf, "%s%s", line.ToRaw(), eol); err != nil {
			// unlikely we will error during writing to a string buffer? maybe we dont need to do anything here
			return err.Error()
		}
	}
	return buf.String()
}

// loadString is a helper function for testing but if we want to expose it some how it's probably safe
func (h *Hosts) loadString(content string) error {
	rdr := strings.NewReader(content)
	scanner := bufio.NewScanner(utfbom.SkipOnly(rdr))
	for scanner.Scan() {
		h.addLine(NewHostsLine(scanner.Text()))
	}
	return scanner.Err()
}

// IsWritable return true if hosts file is writable.
func (h *Hosts) IsWritable() bool {
	file, err := os.OpenFile(h.Path, os.O_WRONLY, 0660)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

// Load the hosts file from the Path into Lines, called by NewHosts() and Hosts.Flush() and you should not need to call this yourself.
func (h *Hosts) Load() error {
	file, err := os.Open(h.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	h.Clear() // reset the lines and lookups in case anything was previously set

	scanner := bufio.NewScanner(utfbom.SkipOnly(file))
	for scanner.Scan() {
		h.addLine(NewHostsLine(scanner.Text()))
	}

	return scanner.Err()
}

// Flush any changes made to hosts file.
func (h *Hosts) Flush() error {
	h.preFlushClean()
	file, err := os.Create(h.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range h.Lines {
		if _, err := fmt.Fprintf(w, "%s%s", line.ToRaw(), eol); err != nil {
			return err
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return h.Load()
}

// AddRaw takes a line from a hosts file and parses/adds the HostsLine
func (h *Hosts) AddRaw(raw ...string) error {
	for _, r := range raw {
		nl := NewHostsLine(r)
		if nl.IP != "" && net.ParseIP(nl.IP) == nil {
			return fmt.Errorf("%q is an invalid IP address", nl.IP)
		}

		for _, host := range nl.Hosts {
			if !govalidator.IsDNSName(host) {
				return fmt.Errorf("hostname is not a valid dns name: %s", host)
			}
		}
		h.addLine(nl)
	}

	return nil
}

// Add an entry to the hosts file.
func (h *Hosts) Add(ip string, hosts ...string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("%q is an invalid IP address", ip)
	}

	// remove hosts from other ips if it already exists
	for _, host := range hosts {
		for _, p := range h.hosts.get(host) {
			if h.Lines[p].IP == ip {
				continue
			}

			if err := h.Remove(h.Lines[p].IP, host); err != nil {
				return err
			}
		}
	}

	position := h.ips.get(ip)
	if len(position) == 0 {
		h.addLine(HostsLine{
			Raw:   fmt.Sprintf("%s %s", ip, strings.Join(hosts, " ")),
			IP:    ip,
			Hosts: hosts,
		})
	} else {
		// add new host to the first one we find
		hostsCopy := h.Lines[position[0]].Hosts
		for _, addHost := range hosts {
			if h.Has(ip, addHost) {
				// this combo already exists
				continue
			}

			if !govalidator.IsDNSName(addHost) {
				return fmt.Errorf("hostname is not a valid dns name: %s", addHost)
			}
			if itemInSliceString(addHost, hostsCopy) {
				continue // host exists for ip already
			}

			hostsCopy = append(hostsCopy, addHost)
			h.hosts.add(addHost, position[0])
		}
		h.Lines[position[0]].Hosts = hostsCopy
		h.Lines[position[0]].Raw = h.Lines[position[0]].ToRaw() // reset raw
	}

	return nil
}

func (h *Hosts) Clear() {
	h.Lines = []HostsLine{}
	h.ips.reset()
	h.hosts.reset()
}

// Clean merge duplicate ips and hosts per ip
func (h *Hosts) Clean() {
	h.CombineDuplicateIPs()
	h.RemoveDuplicateHosts()
	h.SortHosts()
	h.SortByIP()
	h.HostsPerLine(HostsPerLine)
}

// Has return a bool if ip/host combo in hosts file.
func (h *Hosts) Has(ip string, host string) bool {
	ippos := h.ips.get(ip)
	hostpos := h.hosts.get(host)
	for _, pos := range ippos {
		if itemInSliceInt(pos, hostpos) {
			// if ip and host have matching lookup positions we have a combo match
			return true
		}
	}

	return false
}

// HasHostname return a bool if hostname in hosts file.
func (h *Hosts) HasHostname(host string) bool {
	return len(h.hosts.get(host)) > 0
}

// Deprecated: HasIp
func (h *Hosts) HasIp(ip string) bool {
	return h.HasIP(ip)
}

// HasIP will check the
func (h *Hosts) HasIP(ip string) bool {
	return len(h.ips.get(ip)) > 0
}

// Remove takes an ip and an optional host(s), if only an ip is passed the whole line is removed
// when the optional hosts param is passed it will remove only those specific hosts from that ip
func (h *Hosts) Remove(ip string, hosts ...string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("%q is an invalid IP address", ip)
	}

	if len(hosts) == 0 {
		return nil // no point in trying
	}

	lines := make([]HostsLine, len(h.Lines))
	copy(lines, h.Lines)
	h.Clear()

	for _, line := range lines {
		// add back all lines which were not the passed ip
		if line.IP != ip {
			h.addLine(line)
			continue
		}

		var newHosts []string
		for _, checkHost := range line.Hosts {
			if !itemInSliceString(checkHost, hosts) {
				newHosts = append(newHosts, checkHost)
			}
		}

		// If hosts is empty, skip the line completely.
		if len(newHosts) == 0 {
			continue
		}

		// ip still has hosts, add it back
		line.Hosts = newHosts
		line.RegenRaw()

		h.addLine(line)
	}

	return nil
}

// RemoveByHostname go through all lines and remove a hostname if it exists
func (h *Hosts) RemoveByHostname(host string) error {
	restart := true
	for restart {
		restart = false
		for _, p := range h.hosts.get(host) {
			line := &h.Lines[p]
			if len(line.Hosts) > 0 {
				line.Hosts = removeFromSliceString(host, line.Hosts)
				line.RegenRaw()
			}
			h.hosts.remove(host, p)

			// cleanup the whole line if there remains an IP address
			// without hostname/alias
			if len(line.Hosts) == 0 {
				h.removeByPosition(p)
				// when an entry in the lines array is removed
				// the range from hosts.get() above is
				// outdated. Therefore, the whole procedure needs
				// to restart over again
				restart = true
				break
			}
		}
	}

	h.reindex()
	return nil
}

// RemoveByIp this got refactored and wont return an error any more
// leaving it for stable api purposes, will be removed in a major release
func (h *Hosts) RemoveByIp(ip string) error {
	pos := h.ips.get(ip)
	for _, p := range pos {
		h.removeByPosition(p)
	}

	return nil
}

// Deprecated: RemoveDuplicateIps deprecated will be deprecated, use Combine
func (h *Hosts) RemoveDuplicateIps() {
	h.CombineDuplicateIPs()
}

// CombineDuplicateIPs finds all duplicate ips and combines all their hosts into one line
func (h *Hosts) CombineDuplicateIPs() {
	ipCount := make(map[string]int)
	for _, line := range h.Lines {
		if line.IP == "" {
			continue // ignore comments
		}
		ipCount[line.IP]++
	}
	for ip, count := range ipCount {
		if count > 1 {
			h.combineIp(ip)
		}
	}
}

// RemoveDuplicateHosts will check each line and remove hosts if they are the same
func (h *Hosts) RemoveDuplicateHosts() {
	for pos := range h.Lines {
		if h.Lines[pos].IsComment() {
			continue // skip comments
		}
		h.Lines[pos].RemoveDuplicateHosts()
		for _, host := range h.Lines[pos].Hosts {
			h.hosts.remove(host, pos)
		}
	}
}

func (h *Hosts) SortHosts() {
	for pos := range h.Lines {
		h.Lines[pos].SortHosts()
	}
}

// Deprecated: SortByIp switch to SortByIP
func (h *Hosts) SortByIp() {
	h.SortByIP()
}

// SortByIp convert to net.IP and byte.Compare, maintains all comment only lines at the top
func (h *Hosts) SortByIP() {
	// create a new list of unique ips, if dupe ips they will still get grouped together
	sortedIps := make([]net.IP, 0, len(h.Lines))
	unique := make(map[string]struct{})
	for _, l := range h.Lines {
		if _, ok := unique[l.IP]; !ok {
			unique[l.IP] = struct{}{}
			sortedIps = append(sortedIps, net.ParseIP(l.IP))
		}
	}

	// sort the new unique list
	sort.Slice(sortedIps, func(i, j int) bool {
		return bytes.Compare(sortedIps[i], sortedIps[j]) < 0
	})

	// create a copy of of the lines and Clear
	lines := make([]HostsLine, len(h.Lines))
	copy(lines, h.Lines)
	// clear the lines and position indexes to start over
	h.Clear()

	// put all the comments back at the top
	for _, l := range lines {
		if l.IP == "" {
			h.addLine(l)
		}
	}

	// loop over the sorted ips and find their line and add it
	for _, ip := range sortedIps {
		for _, l := range lines {
			if ip.String() == l.IP {
				h.addLine(l) // no continue to group duplicate ips
			}
		}
	}
}

func (h *Hosts) HostsPerLine(count int) {
	// restacks everything into 1 ip again so we can do the split, do this even if count is -1 so it can reset the slice
	if count <= 0 {
		return
	}

	// make a local copy
	lines := make([]HostsLine, len(h.Lines))
	copy(lines, h.Lines)

	// clear the lines and position indexes to start over
	h.Clear()

	for ln, line := range lines {
		if len(line.Hosts) <= count {
			for _, host := range line.Hosts {
				h.hosts.add(host, ln)
			}
			h.ips.add(line.IP, ln)
			h.Lines = append(h.Lines, line)
			continue
		}

		// i: index of the host, j: offset for line number
		for i, j := 0, 0; i < len(line.Hosts); i, j = i+count, j+1 {
			lineCopy := line
			end := len(line.Hosts)
			if end > i+count {
				end = i + count
			}

			for _, host := range line.Hosts {
				h.hosts.add(host, ln+j)
			}
			h.ips.add(line.IP, ln+j)

			lineCopy.Hosts = line.Hosts[i:end]
			lineCopy.Raw = lineCopy.ToRaw()
			h.Lines = append(h.Lines, lineCopy)
		}
	}
}

func (h *Hosts) combineIp(ip string) {
	newLine := HostsLine{
		IP: ip,
	}

	linesCopy := make([]HostsLine, len(h.Lines))
	copy(linesCopy, h.Lines)
	for _, line := range linesCopy {
		if line.IP == ip {
			newLine.Combine(line)
		}
	}
	newLine.SortHosts()
	h.removeIp(ip)
	h.addLine(newLine)
}

// addLine ill append a new HostsLine and add it to the indexes
func (h *Hosts) addLine(line HostsLine) {
	h.Lines = append(h.Lines, line)
	if line.IsComment() {
		return // don't index comments
	}
	pos := len(h.Lines) - 1
	h.ips.add(line.IP, pos)
	for _, host := range line.Hosts {
		h.hosts.add(host, pos)
	}
}

func (h *Hosts) removeByPosition(pos int) {
	if pos == 0 && len(h.Lines) == 1 {
		h.Clear()
		return
	}
	h.Lines = append(h.Lines[:pos], h.Lines[pos+1:]...)
	h.reindex()
}

func (h *Hosts) removeIp(ip string) {
	var newLines []HostsLine
	for _, line := range h.Lines {
		if line.IP != ip {
			newLines = append(newLines, line)
		}
	}

	h.Lines = newLines
	h.reindex()
}

// reindex will reset the internal position arrays for host/ips and rerun the add commands and should be run everytime
// a HostLine is removed. During the add process it's faster to just call the adds instead of reindex as it's more expensive.
func (h *Hosts) reindex() {
	h.hosts.Lock()
	h.hosts.l = make(map[string][]int)
	h.hosts.Unlock()

	h.ips.Lock()
	h.ips.l = make(map[string][]int)
	h.ips.Unlock()

	for pos, line := range h.Lines {
		h.ips.add(line.IP, pos)
		for _, host := range line.Hosts {
			h.hosts.add(host, pos)
		}
	}
}
