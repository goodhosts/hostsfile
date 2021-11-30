package hostsfile

const (
	HostsFilePath = "${SystemRoot}/System32/drivers/etc/hosts"
	HostsPerLine  = 9
	eol           = "\r\n"
)

func (h *Hosts) preFlushClean() {
	// need to force hosts per line always on windows see https://github.com/goodhosts/hostsfile/issues/18
	h.HostsPerLine(HostsPerLine)
}
