package hostsfile

import (
	"fmt"
	"net"
	"strings"
)

type HostsLine struct {
	IP    string
	Hosts []string
	Raw   string
	Err   error
}

const commentChar string = "#"

// Return a new instance of ```HostsLine```.
func NewHostsLine(raw string) HostsLine {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return HostsLine{Raw: raw}
	}

	output := HostsLine{Raw: raw}
	if !output.IsComment() {
		rawIP := fields[0]
		if net.ParseIP(rawIP) == nil {
			output.Err = fmt.Errorf("Bad hosts line: %q", raw)
		}

		output.IP = rawIP
		output.Hosts = fields[1:]
	}

	return output
}

func (l *HostsLine) IsComment() bool {
	return strings.HasPrefix(strings.TrimSpace(l.Raw), commentChar)
}

func (l *HostsLine) IsValid() bool {
	return l.IP != ""
}

func (l *HostsLine) IsMalformed() bool {
	return l.Err != nil
}

func (l *HostsLine) RegenRaw() {
	l.Raw = fmt.Sprintf("%s %s", l.IP, strings.Join(l.Hosts, " "))
}
