package storage

import (
	"sync"
	"testing"
)

func TestNodeShardCacheStorage_GetItem(t *testing.T) {
	type fields struct {
		RWMutex    sync.RWMutex
		lastChange [32]byte
		shardMap   ShardMap
	}
	type args struct {
		lastChange interface{}
		item       interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "WantSuccess",
			fields: fields{
				RWMutex:    sync.RWMutex{},
				lastChange: [32]byte{1, 2, 3, 4, 5, 6, 7},
				shardMap: ShardMap{
					NodeShards: map[int64][]uint64{
						123456789: {1, 3, 5},
					},
					ShardChunks: map[uint64][][]byte{
						1: {
							{
								123, 234,
							},
						},
					},
				},
			},
			args: args{
				lastChange: [32]byte{1, 2, 3, 4, 5, 6, 7},
				item: &ShardMap{
					NodeShards:  nil,
					ShardChunks: nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeShardCacheStorage{
				RWMutex:    tt.fields.RWMutex,
				lastChange: tt.fields.lastChange,
				shardMap:   tt.fields.shardMap,
			}
			if err := n.GetItem(tt.args.lastChange, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.args.item == nil {
				t.Error("GetItem() got nil")
			}
		})
	}
}
