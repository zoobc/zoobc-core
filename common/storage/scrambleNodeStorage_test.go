package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewScrambleCacheStackStorage(t *testing.T) {
	tests := []struct {
		name string
		want *ScrambleCacheStackStorage
	}{
		{
			name: "TestNewScrambleCacheStackStorage:Success",
			want: &ScrambleCacheStackStorage{
				itemLimit:      36,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScrambleCacheStackStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScrambleCacheStackStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScrambleCacheStackStorage_Clear(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_Clear:Success",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.Clear(); (err != nil) != tt.wantErr {
				t.Errorf("Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_GetAll(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		items interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_GetAll:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{},
				},
			},
			args: args{
				items: &[]model.ScrambledNodes{
					{},
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_GetAll:Fail-ItemIsNotScrambleNodes",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				items: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.GetAll(tt.args.items); (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_GetAtIndex(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		index uint32
		item  interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_GetAtIndex:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{},
				},
			},
			args: args{
				index: 0,
				item:  &model.ScrambledNodes{},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_GetAtIndex:Fail-IndexOutOfRange",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				index: 0,
				item:  &model.ScrambledNodes{},
			},
			wantErr: true,
		},
		{
			name: "TestScrambleCacheStackStorage_GetAtIndex:Fail-ItemIsNotScrambleNodes",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{{}},
			},
			args: args{
				index: 0,
				item:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.GetAtIndex(tt.args.index, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetAtIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_GetTop(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_GetTop:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{},
				},
			},
			args: args{
				item: &model.ScrambledNodes{
					IndexNodes:           nil,
					NodePublicKeyToIDMap: nil,
					AddressNodes:         nil,
					BlockHeight:          0,
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_GetTop:Fail-EmptyScramble",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
		{
			name: "TestScrambleCacheStackStorage_GetTop:Fail-ItemIsNotScrambleNode",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{{}},
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.GetTop(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("GetTop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_Pop(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_Pop:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{
						IndexNodes:           nil,
						NodePublicKeyToIDMap: nil,
						AddressNodes:         nil,
						BlockHeight:          0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_Pop:Fail-StackEmpty",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.Pop(); (err != nil) != tt.wantErr {
				t.Errorf("Pop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_PopTo(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		index uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_PopTo:Success",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{
						IndexNodes:           nil,
						NodePublicKeyToIDMap: nil,
						AddressNodes:         nil,
						BlockHeight:          0,
					},
				},
			},
			args: args{
				index: 0,
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_PopTo:Fail-IndexOutOfRange",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				index: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.PopTo(tt.args.index); (err != nil) != tt.wantErr {
				t.Errorf("PopTo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_Push(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		item interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestScrambleCacheStackStorage_Push:Success",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				item: model.ScrambledNodes{
					IndexNodes:           nil,
					NodePublicKeyToIDMap: nil,
					AddressNodes:         nil,
					BlockHeight:          0,
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_Push:Success-Len>0",
			fields: fields{
				itemLimit: 0,
				RWMutex:   sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{
					{},
				},
			},
			args: args{
				item: model.ScrambledNodes{
					IndexNodes:           nil,
					NodePublicKeyToIDMap: nil,
					AddressNodes:         nil,
					BlockHeight:          0,
				},
			},
			wantErr: false,
		},
		{
			name: "TestScrambleCacheStackStorage_Push:Fail-ItemIsNotScrambledNode",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			args: args{
				item: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if err := s.Push(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScrambleCacheStackStorage_copy(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	type args struct {
		src model.ScrambledNodes
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   model.ScrambledNodes
	}{
		{
			name: "TestScrambleCacheStackStorage_copy:Success",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: []model.ScrambledNodes{},
			},
			args: args{
				src: model.ScrambledNodes{
					IndexNodes:           make(map[string]*int),
					NodePublicKeyToIDMap: make(map[string]int64),
					AddressNodes:         make([]*model.Peer, 0),
					BlockHeight:          0,
				},
			},
			want: model.ScrambledNodes{
				IndexNodes:           make(map[string]*int),
				NodePublicKeyToIDMap: make(map[string]int64),
				AddressNodes:         make([]*model.Peer, 0),
				BlockHeight:          0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if got := s.copy(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScrambleCacheStackStorage_size(t *testing.T) {
	type fields struct {
		itemLimit      int
		RWMutex        sync.RWMutex
		scrambledNodes []model.ScrambledNodes
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "TestScrambleCacheStackStorage_size:Success",
			fields: fields{
				itemLimit:      0,
				RWMutex:        sync.RWMutex{},
				scrambledNodes: nil,
			},
			want: 645,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScrambleCacheStackStorage{
				itemLimit:      tt.fields.itemLimit,
				RWMutex:        tt.fields.RWMutex,
				scrambledNodes: tt.fields.scrambledNodes,
			}
			if got := s.size(); got != tt.want {
				t.Errorf("size() = %v, want %v", got, tt.want)
			}
		})
	}
}
