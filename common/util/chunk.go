package util

import (
	"math"
)

type (
	ChunkInterface interface {
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
