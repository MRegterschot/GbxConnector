package lib

import (
	"net"
	"slices"
	"sync"

	"go.uber.org/zap"
)

var (
	cachedLocalIPs []net.IP
	cacheOnce      sync.Once
)

// IsLocalHostname checks if a hostname resolves to a local IP
func IsLocalHostname(hostname string) bool {
	// Shortcut for common local addresses
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return true
	}

	// Load cached local IPs once
	cacheOnce.Do(func() {
		var err error
		cachedLocalIPs, err = getLocalIPs()
		if err != nil {
			zap.L().Error("Failed to get local IPs", zap.Error(err))
			cachedLocalIPs = []net.IP{} // safe empty slice
		}
	})

	// Resolve hostname to IPs
	ips, err := net.LookupIP(hostname)
	if err != nil {
		zap.L().Error("Failed to resolve hostname", zap.String("hostname", hostname), zap.Error(err))
		return false
	}

	// Compare resolved IPs against cached local IPs
	for _, ip := range ips {
		if slices.ContainsFunc(cachedLocalIPs, ip.Equal) {
			return true
		}
	}

	return false
}

// getLocalIPs fetches all local IP addresses, including loopbacks
func getLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // only IPv4
			}
			ips = append(ips, ip)
		}
	}
	// Always include loopback manually
	ips = append(ips, net.ParseIP("127.0.0.1"))
	return ips, nil
}
