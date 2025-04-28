package utils

import (
	"fmt"
	"net"
	"slices"
)

// Checks if a hostname resolves to a local IP
func IsLocalHostname(hostname string) bool {
	// Step 1: Special case for "localhost"
	if hostname == "localhost" {
		return true
	}

	// Step 2: Resolve hostname to IPs
	ips, err := net.LookupIP(hostname)
	if err != nil {
		fmt.Println("Lookup error:", err)
		return false
	}

	// Step 3: Get all local IPs
	localIPs, err := getLocalIPs()
	if err != nil {
		fmt.Println("Local IP fetch error:", err)
		return false
	}

	// Step 4: Compare resolved IPs against local IPs
	for _, ip := range ips {
		if slices.ContainsFunc(localIPs, ip.Equal) {
			return true
		}
	}

	return false
}

// getLocalIPs returns all IP addresses of the local machine
func getLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue // skip this interface
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // skip non-IPv4
			}
			ips = append(ips, ip)
		}
	}
	return ips, nil
}
