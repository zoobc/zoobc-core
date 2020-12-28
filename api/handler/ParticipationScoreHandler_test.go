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
package handler

import (
	"context"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockParticipationScoreService struct {
		service.ParticipationScoreInterface
	}
)

func (*mockParticipationScoreService) GetParticipationScores(
	params *model.GetParticipationScoresRequest) (*model.GetParticipationScoresResponse, error) {
	return &model.GetParticipationScoresResponse{
		ParticipationScores: []*model.ParticipationScore{
			{
				NodeID: 123,
			},
		},
	}, nil
}

func TestParticipationScoreHandler_GetParticipationScores(t *testing.T) {
	type fields struct {
		Service service.ParticipationScoreInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetParticipationScoresRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetParticipationScoresResponse
		wantErr bool
	}{
		{
			name: "wantError:ToHeightIsLowerThanFromHeight",
			fields: fields{
				Service: &mockParticipationScoreService{},
			},
			args: args{
				req: &model.GetParticipationScoresRequest{
					FromHeight: 10,
					ToHeight:   5,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantError:TotalHeightRequestedIsTooMuch",
			fields: fields{
				Service: &mockParticipationScoreService{},
			},
			args: args{
				req: &model.GetParticipationScoresRequest{
					FromHeight: 0,
					ToHeight:   constant.MaxAPILimitPerPage,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				Service: &mockParticipationScoreService{},
			},
			args: args{
				req: &model.GetParticipationScoresRequest{
					FromHeight: 6,
					ToHeight:   10,
				},
			},
			want: &model.GetParticipationScoresResponse{
				ParticipationScores: []*model.ParticipationScore{
					{
						NodeID: 123,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psh := &ParticipationScoreHandler{
				Service: tt.fields.Service,
			}
			got, err := psh.GetParticipationScores(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParticipationScoreHandler.GetParticipationScores() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParticipationScoreHandler.GetParticipationScores() = %v, want %v", got, tt.want)
			}
		})
	}
}
