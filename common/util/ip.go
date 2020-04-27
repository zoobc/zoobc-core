package util

import (
	"errors"
	"net"
)

// GetPublicIP allowing to get own external/public ip,
// Work perfectly on server not on local machine / laptop / PC
// more accurate if getting from request header https://golangcode.com/get-the-request-ip-addr/
func GetPublicIP() (net.IP, error) {
	faces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, face := range faces {
		if face.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if face.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := face.Addrs()
		if err != nil {
			return nil, err
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
				continue // not an ipv4 address
			}
			return ip, nil
		}
	}
	return nil, errors.New("fail caused the internet connection")
}

func IsDomain(address string) bool {
	addr := net.ParseIP(address)
	return addr == nil
}

// IsPublicIP make sure that ip is a public ip or not
func IsPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}
