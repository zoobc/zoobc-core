package auth

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	mockExecutorValidateSuccess struct {
		query.Executor
	}
	nodeAuthMockSignature struct {
		crypto.Signature
		success bool
	}
)

func (navMock *nodeAuthMockSignature) VerifyNodeSignature(payload, signature, nodePublicKey []byte) bool {
	return navMock.success
}

func (*mockExecutorValidateSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockedRows := mock.NewRows(query.NewBlockQuery(&chaintype.MainChain{}).Fields)
	mockedRows.AddRow(
		0,
		[]byte{171, 30, 155, 89, 1, 225, 53, 99, 25, 254, 37, 124, 190, 197, 187, 95, 102, 101, 185, 136, 166, 218, 170,
			156, 49, 43, 208, 228, 157, 166, 224, 91},
		[]byte{},
		1,
		1562806389280,
		[]byte{},
		[]byte{},
		100000000,
		0,
		[]byte{},
		nodePubKey1,
		100000000,
		10000000,
		1,
		0,
	)
	mock.ExpectQuery("SELECT").WillReturnRows(mockedRows)
	return db.QueryRow(qStr), nil
}

func TestProofOfOwnershipValidation_ValidateProofOfOwnership(t *testing.T) {
	poown := GetFixturesProofOfOwnershipValidation(0, nil, nil)
	type fields struct {
		Signature crypto.SignatureInterface
	}
	type args struct {
		poown         *model.ProofOfOwnership
		nodePublicKey []byte
		queryExecutor query.ExecutorInterface
		blockQuery    query.BlockQueryInterface
	}
	poownInvalidSignature := GetFixturesProofOfOwnershipValidation(0, nil, nil)
	poownInvalidSignature.Signature = []byte{41, 7, 108, 68, 19, 119, 1, 128, 65, 227, 181, 177,
		137, 219, 248, 111, 54, 166, 110, 77, 164, 196, 19, 178, 152, 106, 199, 184,
		220, 8, 90, 171, 165, 229, 238, 235, 181, 89, 60, 28, 124, 22, 201, 237, 143,
		63, 59, 156, 133, 194, 189, 97, 150, 245, 96, 45, 192, 236, 109, 80, 14, 31, 243, 10}
	poownInvalidMessage := &model.ProofOfOwnership{
		Signature: []byte{69, 237, 231, 113, 208, 107, 56, 109, 104, 211, 67, 117, 63, 55, 237,
			243, 249, 78, 34, 90, 183, 37, 212, 42, 219, 45, 45, 247, 151, 129, 222, 244, 210,
			185, 54, 184, 17, 214, 72, 231, 195, 159, 171, 184, 73, 193, 84, 224, 51, 37, 139,
			70, 237, 153, 122, 67, 247, 182, 141, 51, 168, 53, 125, 0},
		MessageBytes: []byte{41, 7, 108, 68, 19, 119, 1, 128, 65, 227, 181, 177,
			137, 219, 248, 111, 54, 166, 110, 77, 164, 196, 19, 178, 152, 106, 199, 184,
			220, 8, 90, 171, 165, 229, 238, 235, 181, 89, 60, 28, 124, 22, 201, 237, 143,
			63, 59, 156, 133, 194, 189, 97, 150, 245, 96, 45, 192, 236, 109, 80, 14, 31, 243, 10,
		},
	}
	poownBlockHeightExpired := GetFixturesProofOfOwnershipValidation(101, nil, nil)
	poownBlockInvalidBlockHash := GetFixturesProofOfOwnershipValidation(0, nil, &model.Block{
		ID:                   0,
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		PayloadLength:        0,
		PayloadHash:          []byte{0, 0, 0, 1},
		BlocksmithPublicKey:  nodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	})

	tests := []struct {
		name    string
		p       *NodeAuthValidation
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "Validate:success",
			args: args{
				poown:         poown,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: true},
			},
			wantErr: false,
		},
		{
			name: "Validate:fail-{InvalidSignature}",
			args: args{
				poown:         poownInvalidSignature,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: false},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{InvalidMessageBytes}",
			args: args{
				poown:         poownInvalidMessage,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{BlockHeightExpired}",
			args: args{
				poown:         poownBlockHeightExpired,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: true},
			},
			wantErr: true,
		},
		{
			name: "Validate:fail-{InvalidBlockHash}",
			args: args{
				poown:         poownBlockInvalidBlockHash,
				nodePublicKey: nodePubKey1,
				queryExecutor: &mockExecutorValidateSuccess{},
				blockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: true},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &NodeAuthValidation{
				Signature: tt.fields.Signature,
			}
			err := p.ValidateProofOfOwnership(tt.args.poown, tt.args.nodePublicKey,
				tt.args.queryExecutor, tt.args.blockQuery)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NodeAuthValidation.ValidateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestNodeAuthValidation_ValidateProofOfOrigin(t *testing.T) {
	type fields struct {
		Signature crypto.SignatureInterface
	}
	type args struct {
		poorig            *model.ProofOfOrigin
		nodePublicKey     []byte
		challengeResponse []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ValidateProofOfOrigin:fail-{ProofOfOriginExpired}",
			args: args{
				poorig: &model.ProofOfOrigin{
					MessageBytes: make([]byte, 32),
					Timestamp:    time.Now().Unix() - 1,
					Signature:    []byte{},
				},
				nodePublicKey:     nodePubKey1,
				challengeResponse: make([]byte, 32),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: true},
			},
			wantErr: true,
		},
		{
			name: "ValidateProofOfOrigin:fail-{InvalidChallengeResponse}",
			args: args{
				poorig: &model.ProofOfOrigin{
					MessageBytes: make([]byte, 32),
					Timestamp:    time.Now().Unix() + 10,
					Signature:    []byte{},
				},
				nodePublicKey:     nodePubKey1,
				challengeResponse: make([]byte, 30),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: true},
			},
			wantErr: true,
		},
		{
			name: "ValidateProofOfOrigin:fail-{InvalidSignature}",
			args: args{
				poorig: &model.ProofOfOrigin{
					MessageBytes: make([]byte, 32),
					Timestamp:    time.Now().Unix() + 10,
					Signature:    []byte{},
				},
				nodePublicKey:     nodePubKey1,
				challengeResponse: make([]byte, 32),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: false},
			},
			wantErr: true,
		},
		{
			name: "ValidateProofOfOrigin:{Success}",
			args: args{
				poorig: &model.ProofOfOrigin{
					MessageBytes: make([]byte, 32),
					Timestamp:    time.Now().Unix() + 10,
					Signature:    []byte{},
				},
				nodePublicKey:     nodePubKey1,
				challengeResponse: make([]byte, 32),
			},
			fields: fields{
				Signature: &nodeAuthMockSignature{success: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nav := &NodeAuthValidation{
				Signature: tt.fields.Signature,
			}
			if err := nav.ValidateProofOfOrigin(tt.args.poorig, tt.args.nodePublicKey, tt.args.challengeResponse); (err != nil) != tt.wantErr {
				t.Errorf("NodeAuthValidation.ValidateProofOfOrigin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
