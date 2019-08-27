package service

import (
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

type (
	nodeAdminCoreServiceMocked struct {
		coreService.NodeAdminService
	}
	mockExecutorNodeAdminAPIServiceSuccess struct {
		query.Executor
	}
)

var (
	nodeAdminAPIServicePoown = &model.ProofOfOwnership{
		MessageBytes: []byte{10, 44, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102,
			68, 79, 86, 102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52,
			84, 86, 66, 81, 74, 95, 54, 116, 72, 75, 108, 69, 18, 64, 166, 159, 115, 204,
			162, 58, 154, 197, 200, 181, 103, 220, 24, 90, 117, 110, 151, 201, 130, 22, 79,
			226, 88, 89, 224, 209, 220, 193, 71, 92, 128, 166, 21, 178, 18, 58, 241, 245,
			249, 76, 17, 227, 233, 64, 44, 58, 197, 88, 245, 0, 25, 157, 149, 182, 211, 227,
			1, 117, 133, 134, 40, 29, 205, 38},
		Signature: []byte{0, 0, 0, 0, 41, 7, 108, 68, 19, 119, 1, 128, 65, 227, 181, 177,
			137, 219, 248, 111, 54, 166, 110, 77, 164, 196, 19, 178, 152, 106, 199, 184,
			220, 8, 90, 171, 165, 229, 238, 235, 181, 89, 60, 28, 124, 22, 201, 237, 143,
			63, 59, 156, 133, 194, 189, 97, 150, 245, 96, 45, 192, 236, 109, 80, 14, 31, 243, 10},
	}
	nodeAdminAccountAddress = "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"
)

func (*nodeAdminCoreServiceMocked) GenerateProofOfOwnership(
	nodeAdminAccountAddress string) (*model.ProofOfOwnership, error) {
	return nodeAdminAPIServicePoown, nil
}

func (*nodeAdminCoreServiceMocked) GenerateNodeKey(seed string) ([]byte, error) {
	return []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}, nil
}

func TestNodeAdminService_GetProofOfOwnership(t *testing.T) {
	if err := util.LoadConfig("../resource", "config", "toml"); err != nil {
		logrus.Fatal(err)
	}
	timestamp := time.Now().Unix()
	message := append([]byte(nodeAdminAccountAddress), util.ConvertUint64ToBytes(uint64(timestamp))...)
	sig := crypto.NewSignature().Sign(message, 0, "concur vocalist"+
		" rotten busload gap quote stinging undiluted surfer goofiness deviation starved")

	type fields struct {
		Query                query.ExecutorInterface
		NodeAdminCoreService coreService.NodeAdminServiceInterface
	}
	type args struct {
		nodeAdminAccountAddress string
		timestamp               int64
		signature               []byte
		timeout                 int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.ProofOfOwnership
		wantErr bool
	}{
		{
			name: "GetProofOfOwnership:success",
			fields: fields{
				NodeAdminCoreService: &nodeAdminCoreServiceMocked{},
				Query:                &mockExecutorNodeAdminAPIServiceSuccess{},
			},
			args: args{
				nodeAdminAccountAddress: nodeAdminAccountAddress,
				timestamp:               time.Now().Unix(),
				signature:               sig,
				timeout:                 10,
			},
			want:    nodeAdminAPIServicePoown,
			wantErr: false,
		},
		{
			name: "GetProofOfOwnership:fail-{InvalidTimestamp}",
			fields: fields{
				NodeAdminCoreService: &nodeAdminCoreServiceMocked{},
				Query:                &mockExecutorNodeAdminAPIServiceSuccess{},
			},
			args: args{
				nodeAdminAccountAddress: nodeAdminAccountAddress,
				timestamp:               time.Now().Unix() - viper.GetInt64("proofOfOwnershipReqTimeoutSec"),
				signature:               sig,
				timeout:                 10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				Query:                tt.fields.Query,
				NodeAdminCoreService: tt.fields.NodeAdminCoreService,
			}
			got, err := nas.GetProofOfOwnership(tt.args.nodeAdminAccountAddress, tt.args.timestamp, tt.args.signature, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminService.GetProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminService.GetProofOfOwnership() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdminService_GenerateNodeKey(t *testing.T) {
	if err := util.LoadConfig("../resource", "config", "toml"); err != nil {
		logrus.Fatal(err)
	}
	timestamp := time.Now().Unix()
	message := append([]byte(nodeAdminAccountAddress), util.ConvertUint64ToBytes(uint64(timestamp))...)
	sig := crypto.NewSignature().Sign(message, 0, "concur vocalist"+
		" rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	invalidSig := crypto.NewSignature().Sign(message, 0, "concur vocalist"+
		" rotten busload gap quote stinging undiluted surfer goofiness deviation starved hummus")

	type fields struct {
		Query                query.ExecutorInterface
		NodeAdminCoreService coreService.NodeAdminServiceInterface
	}
	type args struct {
		seed                    string
		nodeAdminAccountAddress string
		timestamp               int64
		signature               []byte
		timeout                 int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:   "GenerateNodeKey:fail-{timestamp expired}",
			fields: fields{},
			args: args{
				timestamp: time.Now().Unix() + 10,
				timeout:   1,
			},
			wantErr: true,
		},
		{
			name: "GenerateNodeKey:success",
			fields: fields{
				NodeAdminCoreService: &nodeAdminCoreServiceMocked{},
				Query:                &mockExecutorNodeAdminAPIServiceSuccess{},
			},
			args: args{
				seed:                    "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness",
				nodeAdminAccountAddress: nodeAdminAccountAddress,
				timestamp:               time.Now().Unix(),
				signature:               sig,
				timeout:                 10,
			},
			want: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
				45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			wantErr: false,
		},
		{
			name: "GenerateNodeKey:success",
			fields: fields{
				NodeAdminCoreService: &nodeAdminCoreServiceMocked{},
				Query:                &mockExecutorNodeAdminAPIServiceSuccess{},
			},
			args: args{
				seed:                    "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness",
				nodeAdminAccountAddress: nodeAdminAccountAddress,
				timestamp:               time.Now().Unix(),
				signature:               invalidSig,
				timeout:                 10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeAdminService{
				Query:                tt.fields.Query,
				NodeAdminCoreService: tt.fields.NodeAdminCoreService,
			}
			got, err := n.GenerateNodeKey(tt.args.nodeAdminAccountAddress, tt.args.timestamp, tt.args.signature, tt.args.timeout, tt.args.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminService.GenerateNodeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminService.GenerateNodeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
