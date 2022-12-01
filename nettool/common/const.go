package common

const IpAddrPattern = "[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+"

const (
	// Localhost ip of localhost
	Localhost = "127.0.0.1"
	// YyyyMmDdHhMmSs timestamp format
	YyyyMmDdHhMmSs = "2006-01-02 15:04:05"
	// StandardSshPort standard ssh port
	StandardSshPort = 22
	// StandardDnsPort standard dns port
	StandardDnsPort = 53
)

const (
	// DnsModeLocalDns local dns mode
	DnsModeLocalDns = "localDNS"
	// DnsModePodDns pod dns mode
	DnsModePodDns = "podDNS"
	// DnsModeHosts hosts dns mode
	DnsModeHosts = "hosts"

	// TunNameWin tun device name in windows
	TunNameWin = "KtConnectTunnel"
	// TunNameLinux tun device name in linux
	TunNameLinux = "kt0"
	// TunNameMac tun device name in MacOS
	TunNameMac = "utun"
	// AlternativeDnsPort alternative port for local dns
	AlternativeDnsPort = 10053
)
