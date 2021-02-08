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
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/storage"
)

var generateRandom32Bytes = func(n int) [][]byte {
	result := make([][]byte, n)
	for i := 0; i < n; i++ {
		rand.Seed(rand.Int63())
		token := make([]byte, 32)
		rand.Read(token)
		result[i] = token
	}
	return result
}
var generateRandomNodeIDs = func(n int) []int64 {
	result := make([]int64, n)
	for i := 0; i < n; i++ {
		result[i] = rand.Int63()
	}
	return result
}

func TestChunk_ShardChunk(t *testing.T) {
	t.Run("6 shardBit-n", func(t *testing.T) {
		const n = 100000
		startPrepareData := time.Now()
		fmt.Printf("preparing %d random data\n", n)
		mockChunks := generateRandom32Bytes(n)
		fmt.Printf("data prepared in: %v ms\n", time.Since(startPrepareData).Milliseconds())
		fmt.Printf("start sharding data\n")
		startSharding := time.Now()
		chunk := &ChunkUtil{
			chunkHashSize: sha256.Size,
		}
		var chunks []byte
		for _, mockChunk := range mockChunks {
			chunks = append(chunks, mockChunk...)
		}
		result := chunk.ShardChunk(chunks, 6)
		fmt.Printf("finish sharding in : %v ms\n", time.Since(startSharding).Milliseconds())
		for u, i := range result {
			fmt.Printf("shardN: %d\tcontent: %d\n", u, len(i))
		}
	})
}

func TestChunk_GetShardAssignment(t *testing.T) {
	t.Run("assignShard - 1000 nodes", func(t *testing.T) {
		const n = 100000
		startPrepareData := time.Now()
		fmt.Printf("preparing %d random data\n", n)
		mockChunks := generateRandom32Bytes(n)
		fmt.Printf("data prepared in: %v ms\n", time.Since(startPrepareData).Milliseconds())
		fmt.Printf("start sharding data\n")
		chunk := &ChunkUtil{
			chunkHashSize:         sha256.Size,
			nodeShardCacheStorage: storage.NewNodeShardCacheStorage(),
			logger:                logrus.New(),
		}
		var chunks []byte
		for _, mockChunk := range mockChunks {
			chunks = append(chunks, mockChunk...)
		}
		nodeIDs := generateRandomNodeIDs(1000)
		startAssignChunk := time.Now()
		shard, err := chunk.GetShardAssignment(chunks, 6, nodeIDs, false)
		if err != nil {
			t.Errorf("error-assigning-shard: %v", err)
		}
		fmt.Printf("time assigning shard: %v ms\n", time.Since(startAssignChunk).Milliseconds())
		for i, s := range shard.NodeShards {
			fmt.Printf("nodeID: %d\tnumShard: %d\n", i, len(s))
		}
	})
}
