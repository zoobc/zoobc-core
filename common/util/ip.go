package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
)

type (
	IPUtil struct {
	}
	IPUtilInterface interface {
		GetPublicIP() (ip *net.IP, err error)
		GetPublicIPDYNDNS() (ip *net.IP, err error)
		IsDomain(address string) bool
		IsPublicIP(ip *net.IP) bool
		DiscoverNodeAddress() (ip *net.IP, err error)
	}
)

// GetPublicIP allowing to get own external/public ip,
// Work perfectly on server not on local machine / laptop / PC
// more accurate if getting from request header https://golangcode.com/get-the-request-ip-addr/
func (ipu *IPUtil) GetPublicIP() (*net.IP, error) {
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
			return &ip, nil
		}
	}
	return nil, errors.New("fail caused the internet connection")
}

// GetPublicIPDYNDNS allowing to get own public ip via http://checkip.dyndns.org
func (ipu *IPUtil) GetPublicIPDYNDNS() (*net.IP, error) {
	var (
		err  error
		bt   []byte
		resp *http.Response
		rgx  = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	)
	resp, err = http.Get("http://checkip.dyndns.org")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bt, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// unfortunately the response is <html> tag, need to get the ip only via regexp
	ipStr := rgx.FindAllString(string(bt), -1)
	ip := net.ParseIP(ipStr[0])
	if ip != nil {
		return &ip, nil
	}
	return nil, fmt.Errorf("invalid ip address")
}

// IsDomain willing to check what kinda address given
func (ipu *IPUtil) IsDomain(address string) bool {
	addr := net.ParseIP(address)
	return addr == nil
}

// IsPublicIP make sure that ip is a public ip or not
func (ipu *IPUtil) IsPublicIP(ip *net.IP) bool {
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

// IsPublicIP make sure that ip is a public ip or not
func (ipu *IPUtil) DiscoverNodeAddress() (ip *net.IP, err error) {
	// first try with an external service
	if ip, err = ipu.GetPublicIPDYNDNS(); err != nil {
		// then locally (and check if discovered address is public)
		if ip, err = ipu.GetPublicIP(); err != nil {
			return nil, err
		}
		if !ipu.IsPublicIP(ip) {
			err = fmt.Errorf("automatically discovered node address %s is not a public IP. "+
				"Your server might be behind a firewall or on a local area network. Note that, "+
				"to be able to actively participate to network activities, generate blocks and keep a high participation score,"+
				"your node must be accessible by other nodes, thus a public IP is required", ip.String())
		}
	}
	return ip, err
}
