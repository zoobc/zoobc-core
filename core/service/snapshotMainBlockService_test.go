package service

import (
	"github.com/pkg/errors"
	"github.com/ugorji/go/codec"
	"os"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockChainType struct {
		chaintype.MainChain
		SnapshotInterval          uint32
		SnapshotGenerationTimeout int64
	}
)

func (mct *mockChainType) GetSnapshotInterval() uint32 {
	return mct.SnapshotInterval
}

func TestSnapshotMainBlockService_IsSnapshotHeight(t *testing.T) {
	type fields struct {
		chainType chaintype.ChainType
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_1}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: 1,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_2}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_3}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 9,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_4}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_5}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 20,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_1}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_2}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_3}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_4}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: 2 * (constant.MinRollbackBlocks + 10),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotMainBlockService{
				chainType: tt.fields.chainType,
			}
			if got := ss.IsSnapshotHeight(tt.args.height); got != tt.want {
				t.Errorf("SnapshotMainBlockService.IsSnapshotHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockFileService struct {
		FileService
		successEncode              bool
		successGetFileNameFromHash bool
		successSaveBytesToFile     bool
	}
	mockSnapshotQueryService struct {
		SnapshotMainBlockQueryService
		successAccountBalances     bool
		successNodeRegistrations   bool
		successAccountDatasets     bool
		successParticipationScores bool
		successPublishedReceipts   bool
		successEscrowTransactions  bool
	}
)

var (
	accBal1 = &model.AccountBalance{
		AccountAddress:   bcsAddress1,
		Balance:          10000000000,
		BlockHeight:      1,
		Latest:           true,
		PopRevenue:       100000000,
		SpendableBalance: 10000000000,
	}
	accBal2 = &model.AccountBalance{
		AccountAddress:   bcsAddress2,
		Balance:          100000000000,
		BlockHeight:      1,
		Latest:           true,
		PopRevenue:       100000000,
		SpendableBalance: 100000000000,
	}
	nr1 = &model.NodeRegistration{
		AccountAddress: bcsAddress1,
		Latest:         true,
		Height:         0,
		LockedBalance:  10000000000,
		NodeAddress: &model.NodeAddress{
			Address: "10.10.10.10",
			Port:    8888,
		},
		NodeID:             11111,
		NodePublicKey:      []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		RegistrationHeight: 0,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
	}
	nr2 = &model.NodeRegistration{
		AccountAddress: bcsAddress2,
		Latest:         true,
		Height:         0,
		LockedBalance:  10000000000,
		NodeAddress: &model.NodeAddress{
			Address: "10.10.10.11",
			Port:    8889,
		},
		NodeID:             22222,
		NodePublicKey:      []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		RegistrationHeight: 0,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
	}
	ps1 = &model.ParticipationScore{
		Latest: true,
		Height: 0,
		NodeID: 11111,
		Score:  1000000,
	}
	pr1 = &model.PublishedReceipt{
		BlockHeight:        1,
		IntermediateHashes: make([]byte, 32),
		PublishedIndex:     100,
		ReceiptIndex:       10,
	}
	escrowTx1 = &model.Escrow{
		BlockHeight:      1,
		Latest:           true,
		ID:               999999,
		Amount:           1000000000,
		ApproverAddress:  bcsAddress1,
		Commission:       100000000,
		Instruction:      "test test",
		RecipientAddress: bcsAddress2,
		SenderAddress:    bcsAddress3,
		Status:           model.EscrowStatus_Pending,
		Timeout:          15875392,
	}
	accDataSet1 = &model.AccountDataset{
		Height:                  1,
		Latest:                  true,
		Property:                "testProp",
		RecipientAccountAddress: bcsAddress1,
		SetterAccountAddress:    bcsAddress2,
		TimestampExpires:        15875392,
		TimestampStarts:         15875000,
		Value:                   "testVal",
	}
	blockForSnapshot1 = &model.Block{
		Height:    720,
		Timestamp: 15875392,
	}
	snapshotFullHash = []byte{
		189, 123, 189, 67, 77, 99, 212, 229, 139, 70, 138, 166, 32, 117, 190, 42, 156, 137, 6, 216, 156, 116, 20, 182, 211, 178,
		224, 220, 235, 28, 62, 12,
	}
)

func (*mockFileService) HashPayload(b []byte) []byte {
	return snapshotFullHash
}

func (mfs *mockFileService) EncodePayload(v interface{}) (b []byte, err error) {
	b = []byte{
		130, 166, 110, 65, 99, 99, 111, 117, 110, 116, 65, 100, 100, 114, 101, 115, 115, 120, 44, 66, 67, 90, 110, 83, 102,
		113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89, 107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118,
		76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116, 78, 103, 66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 2, 84, 11,
		228, 0, 107, 66, 108, 111, 99, 107, 72, 101, 105, 103, 104, 116, 1, 102, 76, 97, 116, 101, 115, 116, 245, 106, 80,
		111, 112, 82, 101, 118, 101, 110, 117, 101, 26, 5, 245, 225, 0, 112, 83, 112, 101, 110, 100, 97, 98, 108, 101, 66,
		97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 2, 84, 11, 228, 0, 166, 110, 65, 99, 99, 111, 117, 110, 116, 65, 100, 100,
		114, 101, 115, 115, 120, 44, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75,
		111, 74, 115, 107, 106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 103,
		66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 23, 72, 118, 232, 0, 107, 66, 108, 111, 99, 107, 72, 101, 105, 103,
		104, 116, 1, 102, 76, 97, 116, 101, 115, 116, 245, 106, 80, 111, 112, 82, 101, 118, 101, 110, 117, 101, 26, 5, 245,
		225, 0, 112, 83, 112, 101, 110, 100, 97, 98, 108, 101, 66, 97, 108, 97, 110, 99, 101, 27, 0, 0, 0, 23, 72, 118,
		232, 0,
	}
	if mfs.successEncode {
		return b, nil
	}
	return nil, errors.New("EncodedPayloadFail")
}

func (mfs *mockFileService) GetFileNameFromHash(fileHash []byte) (string, error) {
	if mfs.successGetFileNameFromHash {
		return "vXu9Q01j1OWLRoqmIHW-KpyJBticdBS207Lg3OscPgyO", nil
	}
	return "", errors.New("GetFileNameFromHashFail")
}
func (mfs *mockFileService) SaveBytesToFile(filePath string, b []byte) (*os.File, error) {
	if mfs.successSaveBytesToFile {
		return nil, nil
	}
	return nil, errors.New("SaveBytesToFileFail")
}

func (msqs *mockSnapshotQueryService) GetAccountBalances(fromHeight, toHeight uint32) ([]*model.AccountBalance, error) {
	if msqs.successAccountBalances {
		return []*model.AccountBalance{
			accBal1,
			accBal2,
		}, nil
	}
	return nil, errors.New("GetAccountBalancesFail")
}

func (msqs *mockSnapshotQueryService) GetNodeRegistrations(fromHeight, toHeight uint32) ([]*model.NodeRegistration, error) {
	if msqs.successNodeRegistrations {
		return []*model.NodeRegistration{
			nr1,
			nr2,
		}, nil
	}
	return nil, errors.New("GetNodeRegistrationsFail")
}

func (msqs *mockSnapshotQueryService) GetAccountDatasets(fromHeight, toHeight uint32) ([]*model.AccountDataset, error) {
	if msqs.successAccountDatasets {
		return []*model.AccountDataset{
			accDataSet1,
		}, nil
	}
	return nil, errors.New("GetAccountDatasetsFail")
}

func (msqs *mockSnapshotQueryService) GetParticipationScores(fromHeight, toHeight uint32) ([]*model.ParticipationScore, error) {
	if msqs.successParticipationScores {
		return []*model.ParticipationScore{
			ps1,
		}, nil
	}
	return nil, errors.New("GetParticipationScoresFail")
}

func (msqs *mockSnapshotQueryService) GetPublishedReceipts(fromHeight, toHeight, limit uint32) ([]*model.PublishedReceipt, error) {
	if msqs.successPublishedReceipts {
		return []*model.PublishedReceipt{
			pr1,
		}, nil
	}
	return nil, errors.New("GetPublishedReceiptsFail")
}

func (msqs *mockSnapshotQueryService) GetEscrowTransactions(fromHeight, toHeight uint32) ([]*model.Escrow, error) {
	if msqs.successEscrowTransactions {
		return []*model.Escrow{
			escrowTx1,
		}, nil
	}
	return nil, errors.New("GetEscrowTransactionsFail")
}

func (mct *mockChainType) GetSnapshotGenerationTimeout() int64 {
	return mct.SnapshotGenerationTimeout
}

func TestSnapshotMainBlockService_NewSnapshotFile(t *testing.T) {
	type fields struct {
		SnapshotPath string
		chainType    chaintype.ChainType
		Logger       *log.Logger
		QueryService SnapshotMainBlockQueryServiceInterface
		FileService  FileServiceInterface
	}
	type args struct {
		block          *model.Block
		chunkSizeBytes int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.SnapshotFileInfo
		wantErr bool
		errMsg  string
	}{
		{
			name: "NewSnapshotFile:success",
			fields: fields{
				FileService: &mockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
					successEncode:              true,
					successGetFileNameFromHash: true,
					successSaveBytesToFile:     true,
				},
				Logger:       log.New(),
				SnapshotPath: "testdata/snapshots",
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:     true,
					successNodeRegistrations:   true,
					successAccountDatasets:     true,
					successParticipationScores: true,
					successPublishedReceipts:   true,
					successEscrowTransactions:  true,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want: &model.SnapshotFileInfo{
				SnapshotFileHash: snapshotFullHash,
				FileChunksHashes: [][]byte{
					snapshotFullHash,
				},
				ChainType:                  0,
				Height:                     blockForSnapshot1.Height,
				ProcessExpirationTimestamp: blockForSnapshot1.Timestamp + 1,
				SpineBlockManifestType:     model.SpineBlockManifestType_Snapshot,
			},
		},
		{
			name: "NewSnapshotFile:fail-{GetAccountBalances}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances: false,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "GetAccountBalancesFail",
		},
		{
			name: "NewSnapshotFile:fail-{GetNodeRegistrations}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:   true,
					successNodeRegistrations: false,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "GetNodeRegistrationsFail",
		},
		{
			name: "NewSnapshotFile:fail-{GetAccountDatasets}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:   true,
					successNodeRegistrations: true,
					successAccountDatasets:   false},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "GetAccountDatasetsFail",
		},
		{
			name: "NewSnapshotFile:fail-{GetParticipationScores}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:     true,
					successNodeRegistrations:   true,
					successAccountDatasets:     true,
					successParticipationScores: false,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "GetParticipationScoresFail",
		},
		{
			name: "NewSnapshotFile:fail-{GetPublishedRecepits}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:     true,
					successNodeRegistrations:   true,
					successAccountDatasets:     true,
					successParticipationScores: true,
					successPublishedReceipts:   false,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "GetPublishedReceiptsFail",
		},
		{
			name: "NewSnapshotFile:fail-{GetEscrowTransactions}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:     true,
					successNodeRegistrations:   true,
					successAccountDatasets:     true,
					successParticipationScores: true,
					successPublishedReceipts:   true,
					successEscrowTransactions:  false,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "GetEscrowTransactionsFail",
		},
		{
			name: "NewSnapshotFile:fail-{EncodedPayload}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				FileService: &mockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
					successEncode:              false,
					successGetFileNameFromHash: true,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:     true,
					successNodeRegistrations:   true,
					successAccountDatasets:     true,
					successParticipationScores: true,
					successPublishedReceipts:   true,
					successEscrowTransactions:  true,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "EncodedPayloadFail",
		},
		{
			name: "NewSnapshotFile:fail-{GetFileNameFromHash}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				FileService: &mockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
					successEncode:              true,
					successGetFileNameFromHash: false,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:     true,
					successNodeRegistrations:   true,
					successAccountDatasets:     true,
					successParticipationScores: true,
					successPublishedReceipts:   true,
					successEscrowTransactions:  true,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "GetFileNameFromHashFail",
		},
		{
			name: "NewSnapshotFile:fail-{SaveBytesToFile}",
			fields: fields{
				chainType: &mockChainType{
					SnapshotGenerationTimeout: 1,
				},
				FileService: &mockFileService{
					FileService: FileService{
						Logger: log.New(),
						h:      new(codec.CborHandle),
					},
					successEncode:              true,
					successGetFileNameFromHash: true,
					successSaveBytesToFile:     false,
				},
				QueryService: &mockSnapshotQueryService{
					successAccountBalances:     true,
					successNodeRegistrations:   true,
					successAccountDatasets:     true,
					successParticipationScores: true,
					successPublishedReceipts:   true,
					successEscrowTransactions:  true,
				},
			},
			args: args{
				block: blockForSnapshot1,
			},
			want:    nil,
			wantErr: true,
			errMsg:  "SaveBytesToFileFail",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotMainBlockService{
				SnapshotPath: tt.fields.SnapshotPath,
				chainType:    tt.fields.chainType,
				Logger:       tt.fields.Logger,
				QueryService: tt.fields.QueryService,
				FileService:  tt.fields.FileService,
			}
			got, err := ss.NewSnapshotFile(tt.args.block, tt.args.chunkSizeBytes)
			if err != nil {
				if tt.wantErr {
					if tt.errMsg != err.Error() {
						t.Errorf("error differs from what expected. wrong test exit line. gotErr %s, wantErr %s",
							err.Error(),
							tt.errMsg)
					}
					return
				}
				t.Errorf("SnapshotMainBlockService.NewSnapshotFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotMainBlockService.NewSnapshotFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
