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
