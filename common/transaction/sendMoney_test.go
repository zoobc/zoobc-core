package transaction

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestSendMoney_Validate(t *testing.T) {
	type fields struct {
		Body                 *model.SendMoneyTransactionBody
		SenderAddress        string
		SenderAccountType    uint32
		RecipientAddress     string
		RecipientAccountType uint32
		Height               uint32
		AccountBalanceQuery  query.AccountBalanceInt
		AccountQuery         query.AccountQueryInterface
		QueryExecutor        query.ExecutorInterface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: 10,
				},
				Height: 0,
			},
			wantErr: false,
		},
		{
			name: "wantError",
			fields: fields{
				Body: &model.SendMoneyTransactionBody{
					Amount: -1,
				},
				Height: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SendMoney{
				Body:                 tt.fields.Body,
				SenderAddress:        tt.fields.SenderAddress,
				SenderAccountType:    tt.fields.SenderAccountType,
				RecipientAddress:     tt.fields.RecipientAddress,
				RecipientAccountType: tt.fields.RecipientAccountType,
				Height:               tt.fields.Height,
				AccountBalanceQuery:  tt.fields.AccountBalanceQuery,
				AccountQuery:         tt.fields.AccountQuery,
				QueryExecutor:        tt.fields.QueryExecutor,
			}
			if err := tx.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("SendMoney.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
