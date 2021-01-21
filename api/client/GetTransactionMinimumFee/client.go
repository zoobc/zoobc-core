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
package main

import (
	"context"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var (
		ip string
	)
	flag.StringVar(&ip, "ip", "", "Usage")
	flag.Parse()
	if len(ip) < 1 {
		if err := util.LoadConfig("../../../", "config", "toml", ""); err != nil {
			log.Fatal(err)
		} else {
			ip = fmt.Sprintf(":%d", viper.GetInt("apiRPCPort"))
		}
	}

	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewTransactionServiceClient(conn)

	request := &rpc_model.GetTransactionMinimumFeeRequest{
		TransactionBytes: []byte{
			1, 0, 0, 0, 1, 191, 62, 75, 94, 0, 0, 0, 0, 44, 0, 0, 0, 122, 85, 107, 50, 109, 99, 103, 57, 118,
			110, 81, 115, 102, 79, 111, 110, 49, 84, 45, 51, 102, 108, 55, 56, 80, 78, 106, 100, 109, 68, 55,
			49, 66, 54, 86, 99, 45, 101, 72, 65, 56, 102, 79, 54, 44, 0, 0, 0, 79, 110, 69, 89, 122, 73, 45,
			69, 77, 86, 54, 85, 84, 102, 111, 85, 69, 122, 112, 81, 85, 106, 107, 83, 108, 110, 113, 66, 56,
			50, 45, 83, 121, 82, 78, 55, 52, 54, 57, 108, 74, 84, 87, 72, 1, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0,
			100, 0, 0, 0, 0, 0, 0, 0, 44, 0, 0, 0, 79, 110, 69, 89, 122, 73, 45, 69, 77, 86, 54, 85, 84, 102,
			111, 85, 69, 122, 112, 81, 85, 106, 107, 83, 108, 110, 113, 66, 56, 50, 45, 83, 121, 82, 78, 55,
			52, 54, 57, 108, 74, 84, 87, 72, 1, 0, 0, 0, 0, 0, 0, 0, 141, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 151, 14, 69, 104, 198, 182, 113, 50, 53, 190, 105, 20, 164, 57, 16, 94, 89, 251, 35, 230,
			145, 198, 189, 167, 222, 214, 208, 120, 229, 172, 155, 54, 85, 245, 19, 125, 3, 4, 11, 44, 65, 254,
			148, 174, 117, 98, 16, 161, 149, 16, 4, 0, 153, 107, 84, 187, 8, 225, 103, 208, 126, 101, 17, 0,
		},
	}
	response, err := c.GetTransactionMinimumFee(context.Background(), request)

	if err != nil {
		log.Fatalf("error calling rpc_service.GetTransactionMinimumFee: %s", err)
	}

	log.Printf("response from remote rpc_service.GetTransactionMinimumFee(%v): %v", request, response.GetFee())

}
