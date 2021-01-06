// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
