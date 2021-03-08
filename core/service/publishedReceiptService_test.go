// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package service

import (
	"errors"
	"github.com/zoobc/zoobc-core/common/constant"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/util"
)

func TestNewPublishedReceiptService(t *testing.T) {
	type args struct {
		publishedReceiptQuery query.PublishedReceiptQueryInterface
		receiptUtil           util.ReceiptUtilInterface
		publishedReceiptUtil  util.PublishedReceiptUtilInterface
		receiptService        ReceiptServiceInterface
		queryExecutor         query.ExecutorInterface
	}
	tests := []struct {
		name string
		args args
		want *PublishedReceiptService
	}{
		{
			name: "NewPublishedReceiptService-Success",
			args: args{
				publishedReceiptQuery: nil,
				receiptUtil:           nil,
				publishedReceiptUtil:  nil,
				receiptService:        nil,
				queryExecutor:         nil,
			},
			want: &PublishedReceiptService{
				PublishedReceiptQuery: nil,
				ReceiptUtil:           nil,
				PublishedReceiptUtil:  nil,
				ReceiptService:        nil,
				QueryExecutor:         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPublishedReceiptService(tt.args.publishedReceiptQuery, tt.args.receiptUtil, tt.args.publishedReceiptUtil,
				tt.args.receiptService, tt.args.queryExecutor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPublishedReceiptService() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	// ProcessPublishedReceipts mocks
	mockProcessPublishedReceiptsReceiptServiceFail struct {
		ReceiptService
	}
	mockProcessPublishedReceiptsReceiptServiceSuccess struct {
		ReceiptService
	}
	mockProcessPublishedReceiptPublishedReceiptUtilFail struct {
		util.PublishedReceiptUtil
	}
	mockProcessPublishedReceiptPublishedReceiptUtilSuccess struct {
		util.PublishedReceiptUtil
	}
	// ProcessPublishedReceipts mocks
)

func (*mockProcessPublishedReceiptsReceiptServiceFail) ValidateReceipt(
	_ *model.Receipt,
	_ bool,
) error {
	return errors.New("mockedError")
}

func (*mockProcessPublishedReceiptsReceiptServiceSuccess) ValidateReceipt(
	_ *model.Receipt,
	_ bool,
) error {
	return nil
}

func (*mockProcessPublishedReceiptsReceiptServiceSuccess) ValidateUnlinkedReceipts(
	receiptsToValidate []*model.PublishedReceipt,
	blockToValidate *model.Block,
) (validReceipts []*model.PublishedReceipt, err error) {
	return mockPublishedReceipt, nil
}

func (*mockProcessPublishedReceiptsReceiptServiceFail) ValidateUnlinkedReceipts(
	receiptsToValidate []*model.PublishedReceipt,
	blockToValidate *model.Block,
) (validReceipts []*model.PublishedReceipt, err error) {
	return []*model.PublishedReceipt{}, nil
}
func (*mockProcessPublishedReceiptsReceiptServiceSuccess) ValidateLinkedReceipts(
	receiptsToValidate []*model.PublishedReceipt,
	blockToValidate *model.Block,
	maxLookBackwardSteps int32,
) (validReceipts []*model.PublishedReceipt, err error) {
	return mockPublishedReceipt, nil
}

func (*mockProcessPublishedReceiptsReceiptServiceFail) ValidateLinkedReceipts(
	receiptsToValidate []*model.PublishedReceipt,
	blockToValidate *model.Block,
	maxLookBackwardSteps int32,
) (validReceipts []*model.PublishedReceipt, err error) {
	return []*model.PublishedReceipt{}, nil
}

func (*mockProcessPublishedReceiptPublishedReceiptUtilFail) InsertPublishedReceipt(
	_ *model.PublishedReceipt, _ bool,
) error {
	return errors.New("mockedError")
}

func (*mockProcessPublishedReceiptPublishedReceiptUtilSuccess) InsertPublishedReceipt(
	_ *model.PublishedReceipt, _ bool,
) error {
	return nil
}

func TestPublishedReceiptService_ProcessPublishedReceipts(t *testing.T) {
	dummyPublishedReceipts := []*model.PublishedReceipt{
		{
			Receipt: &model.Receipt{},
		},
	}
	type fields struct {
		PublishedReceiptQuery   query.PublishedReceiptQueryInterface
		ReceiptUtil             util.ReceiptUtilInterface
		PublishedReceiptUtil    util.PublishedReceiptUtilInterface
		ReceiptService          ReceiptServiceInterface
		QueryExecutor           query.ExecutorInterface
		PublishedReceiptService PublishedReceiptServiceInterface
	}
	type args struct {
		block                     *model.Block
		numberOfEstimatedReceipts uint32
		validateReceipt           bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "ProcessPublishedReceipt-NoReceipt",
			fields: fields{
				PublishedReceiptQuery: nil,
				ReceiptUtil:           nil,
				PublishedReceiptUtil:  nil,
				ReceiptService:        &mockProcessPublishedReceiptsReceiptServiceFail{},
				QueryExecutor:         nil,
			},
			args: args{
				block: &model.Block{
					PublishedReceipts: make([]*model.PublishedReceipt, 0),
					Height:            constant.BatchReceiptLookBackHeight,
				},
				numberOfEstimatedReceipts: 5,
				validateReceipt:           true,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "ProcessPublishedReceipt-ValidateReceiptSuccess",
			fields: fields{
				PublishedReceiptQuery: nil,
				ReceiptUtil:           nil,
				PublishedReceiptUtil:  &mockProcessPublishedReceiptPublishedReceiptUtilSuccess{},
				ReceiptService:        &mockProcessPublishedReceiptsReceiptServiceSuccess{},
				QueryExecutor:         nil,
			},
			args: args{
				block: &model.Block{
					PublishedReceipts: dummyPublishedReceipts,
					Height:            constant.BatchReceiptLookBackHeight,
				},
				numberOfEstimatedReceipts: 5,
				validateReceipt:           true,
			},
			want:    len(mockPublishedReceipt),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &PublishedReceiptService{
				PublishedReceiptQuery: tt.fields.PublishedReceiptQuery,
				ReceiptUtil:           tt.fields.ReceiptUtil,
				PublishedReceiptUtil:  tt.fields.PublishedReceiptUtil,
				ReceiptService:        tt.fields.ReceiptService,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			got, _, err := ps.ProcessPublishedReceipts(tt.args.block, tt.args.numberOfEstimatedReceipts, tt.args.validateReceipt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessPublishedReceipts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProcessPublishedReceipts() got = %v, want %v", got, tt.want)
			}
		})
	}
}
