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
