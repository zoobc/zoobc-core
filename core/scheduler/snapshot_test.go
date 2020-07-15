package scheduler

import (
	"database/sql"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	mockSpineBlockManifestErr struct {
		service.SpineBlockManifestService
	}
	mockSpineBlockManifestSuccess struct {
		service.SpineBlockManifestService
	}
)

func (*mockSpineBlockManifestErr) GetLastSpineBlockManifest(
	chaintype.ChainType,
	model.SpineBlockManifestType,
) (sm *model.SpineBlockManifest, err error) {
	return nil, sql.ErrNoRows
}
func (*mockSpineBlockManifestSuccess) GetLastSpineBlockManifest(
	chaintype.ChainType,
	model.SpineBlockManifestType,
) (sm *model.SpineBlockManifest, err error) {

	payload := model.SnapshotPayload{
		Blocks: []*model.Block{transaction.GetFixturesForBlock(720, 1234567890)},
	}

	fullHashed, fileChunkHashes, err := service.NewSnapshotBasicChunkStrategy(constant.SnapshotChunkSize, service.NewFileService(
		logrus.New(),
		new(codec.CborHandle),
		"./testdata",
	)).GenerateSnapshotChunks(&payload)
	if err != nil {
		return nil, err
	}
	var fullChunkHashesx []byte
	for _, chunkHaHash := range fileChunkHashes {
		fullChunkHashesx = append(fullChunkHashesx, chunkHaHash...)
	}
	return &model.SpineBlockManifest{
		ID:                       12345678,
		FullFileHash:             fullHashed,
		FileChunkHashes:          fullChunkHashesx,
		ManifestReferenceHeight:  720,
		ManifestSpineBlockHeight: 1,
		ChainType:                0,
		SpineBlockManifestType:   0,
		ExpirationTimestamp:      567890,
	}, nil
}
func TestSnapshotScheduler_CheckChunksIntegrity(t *testing.T) {
	type fields struct {
		SpineBlockManifestService service.SpineBlockManifestServiceInterface
		FileService               service.FileServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		filePath  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "WantErr:SpineBlockManifest",
			fields: fields{
				SpineBlockManifestService: &mockSpineBlockManifestErr{},
			},
			wantErr: true,
		},
		{
			name: "WantErr:SpineBlockManifestParseFail",
			fields: fields{
				SpineBlockManifestService: &mockSpineBlockManifestSuccess{},
				FileService: service.NewFileService(
					logrus.New(),
					new(codec.CborHandle),
					"./testdata",
				),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotScheduler{
				SpineBlockManifestService: tt.fields.SpineBlockManifestService,
				FileService:               tt.fields.FileService,
			}
			if err := ss.CheckChunksIntegrity(tt.args.chainType, tt.args.filePath); err != nil && !tt.wantErr {
				t.Errorf("CheckChunksIntegrity got err: %s", err.Error())
			}
		})
	}
}
