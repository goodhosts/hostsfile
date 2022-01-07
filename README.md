# Go library for working with a system's hostsfile
[![codecov](https://codecov.io/gh/goodhosts/hostsfile/branch/master/graph/badge.svg?token=BJQH16QQEH)](https://codecov.io/gh/goodhosts/hostsfile)
## Usage

Using system default hosts file
```
hfile, err := hostsfile.NewHosts()
```

Using a custom hostsfile at a specific location
```
hfile, err := hostsfile.NewCustomHosts("./my-custom-hostsfile")
```

Add an ip entry with it's hosts
```
err := hfile.Add("192.168.1.1", "my-hostname", "another-hostname")
```

Remove an ip/host combination
```
err := hfile.Remove("192.168.1.1", "another-hostname")
```

Flush the hostfile changes back to disk
```
err := hfile.Flush()
```

# Full API
```
type Hosts
func NewCustomHosts(osHostsFilePath string) (*Hosts, error)
    func NewHosts() (*Hosts, error)
    func (h *Hosts) Add(ip string, hosts ...string) error
    func (h *Hosts) AddRaw(raw ...string) error
    func (h *Hosts) Clean()
    func (h *Hosts) Clear()
    func (h *Hosts) Flush() error
    func (h *Hosts) Has(ip string, host string) bool
    func (h *Hosts) HasHostname(host string) bool
    func (h *Hosts) HasIp(ip string) bool
    func (h *Hosts) HostsPerLine(count int)
    func (h *Hosts) IsWritable() bool
    func (h *Hosts) Load() error
    func (h *Hosts) Remove(ip string, hosts ...string) error
    func (h *Hosts) RemoveByHostname(host string) error
    func (h *Hosts) RemoveByIp(ip string) error
    func (h *Hosts) RemoveDuplicateHosts()
    func (h *Hosts) RemoveDuplicateIps()
    func (h *Hosts) SortByIp()
    func (h *Hosts) SortHosts()
type HostsLine
func NewHostsLine(raw string) HostsLine
    func (l *HostsLine) Combine(hostline HostsLine)
    func (l *HostsLine) HasComment() bool
    func (l *HostsLine) IsComment() bool
    func (l *HostsLine) IsMalformed() bool
    func (l *HostsLine) IsValid() bool
    func (l *HostsLine) RegenRaw()
    func (l *HostsLine) RemoveDuplicateHosts()
    func (l *HostsLine) SortHosts()
    func (l *HostsLine) ToRaw() string
```