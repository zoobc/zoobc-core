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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/crypto"

	log "github.com/sirupsen/logrus"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	var ip string
	flag.StringVar(&ip, "ip", "", "Usage")
	flag.Parse()
	if len(ip) < 1 {
		config, err := util.LoadConfig("../../../", "config", "toml", "", false)
		if err != nil {
			log.Fatal(err)
		} else {
			ip = fmt.Sprintf(":%d", config.RPCAPIPort)
		}
	}
	conn, err := grpc.Dial(ip, grpc.WithInsecure())

	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewNodeAdminServiceClient(conn)

	signature := crypto.Signature{}
	accountSeed := "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
	currentTime := uint64(time.Now().Unix())
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(currentTime))
	buffer.Write(util.ConvertUint32ToBytes(uint32(rpcModel.RequestType_GeneratetNodeKey)))
	sig, err := signature.Sign(
		buffer.Bytes(),
		rpcModel.AccountType_ZbcAccountType,
		accountSeed,
	)
	if err != nil {
		log.Fatalf("error signing payload: %s", err)
	}
	buffer.Write(sig)
	ctx := context.Background()
	md := metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(buffer.Bytes()))
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := c.GenerateNodeKey(ctx, &rpcModel.GenerateNodeKeyRequest{})

	if err != nil {
		log.Fatalf("error calling remote.GenerateNodeKey: %s", err)
	}
	j, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("response from remote.GenerateNodeKey(): %s", j)
}
