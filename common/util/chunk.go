package util

import (
	"math"
	"math/big"
	"sort"

	"golang.org/x/crypto/sha3"
)

type (
	ChunkInterface interface {
		ShardChunk(chunks []byte, shardBitLength int) map[uint64][][]byte
	}

	Chunk struct {
		chunkSize int
	}
)

// ShardChunk accept chunks and number of shard identification bits
// return the mapped chunks to their respective shard
func (c *Chunk) ShardChunk(chunks []byte, shardBitLength int) map[uint64][][]byte {
	var (
		shards  = make(map[uint64][][]byte)
		bitMask = (1 << shardBitLength) - 1
	)
	shardByteLength := int(math.Ceil(float64(shardBitLength) / 8))
	byteMasking := make([]byte, 8-shardByteLength)
	for i := 0; i < len(chunks); i += c.chunkSize {
		var (
			chunkShardByte = make([]byte, c.chunkSize)
			chunkByte      = make([]byte, c.chunkSize)
		)
		// check if chunkShardByte in which shard
		copy(chunkByte, chunks[i:i+c.chunkSize])          // prepare copy of chunk
		copy(chunkShardByte, chunks[i:i+shardByteLength]) // prepare a copy of the shard identity slice
		chunkShardByte = append(chunkShardByte, byteMasking...)
		shardByteUint64 := ConvertBytesToUint64(chunkShardByte)
		shardNumber := shardByteUint64 & uint64(bitMask) // msb masking
		shards[shardNumber] = append(shards[shardNumber], chunks[i:i+c.chunkSize])
	}
	return shards
}

// AssignShard assign built shard to provided nodeIDs and return the mapped data + cache to CacheStorage
func (c *Chunk) AssignShard(
	shards map[uint64][][]byte,
	nodeIDs []int64,
) (map[int64][]uint64, error) {
	type nodeOrder struct {
		nodeID int64
		hash   []byte
	}
	var (
		nodeShards      = make(map[int64][]uint64)
		shardRedundancy = int(math.Ceil(math.Sqrt(float64(len(nodeIDs)))))
	)
	for shardNumber := range shards {
		var nodeOrders = make([]nodeOrder, len(nodeIDs))
		for i := 0; i < len(nodeIDs); i++ { // todo: split hashing to multiple goroutines
			digest := sha3.New256()
			if _, err := digest.Write(ConvertUint64ToBytes(uint64(nodeIDs[i]))); err != nil {
				return nil, err
			}
			if _, err := digest.Write(ConvertUint64ToBytes(shardNumber)); err != nil {
				return nil, err
			}
			nodeOrders[i] = nodeOrder{
				nodeID: nodeIDs[i],
				hash:   digest.Sum([]byte{}),
			}
		}
		// sort nodeOrders
		sort.SliceStable(nodeOrders, func(i, j int) bool {
			resI := new(big.Int).SetBytes(nodeOrders[i].hash)
			resJ := new(big.Int).SetBytes(nodeOrders[j].hash)
			res := resI.Cmp(resJ)
			// Ascending sort
			return res < 0
		})
		for i := 0; i < shardRedundancy; i++ {
			nodeShards[nodeOrders[i].nodeID] = append(nodeShards[nodeOrders[i].nodeID], shardNumber)
		}
	}
	// todo: cache nodeShards
	return nodeShards, nil
}
