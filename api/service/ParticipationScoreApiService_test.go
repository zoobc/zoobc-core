package service

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

type (
	mockParticipationScoreService struct {
		coreService.ParticipationScoreServiceInterface
	}
)

func (*mockParticipationScoreService) GetParticipationScoreByBlockHeightRange(fromBlockHeight,
	toBlockHeight uint32) ([]*model.ParticipationScore, error) {
	return []*model.ParticipationScore{
		{
			NodeID: 123,
		},
	}, nil
}

func TestParticipationScoreService_GetParticipationScores(t *testing.T) {
	type fields struct {
		ParticipationScoreService coreService.ParticipationScoreServiceInterface
	}
	type args struct {
		params *model.GetParticipationScoresRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetParticipationScoresResponse
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				ParticipationScoreService: &mockParticipationScoreService{},
			},
			args: args{
				params: &model.GetParticipationScoresRequest{},
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
			pss := &ParticipationScoreService{
				ParticipationScoreService: tt.fields.ParticipationScoreService,
			}
			got, err := pss.GetParticipationScores(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParticipationScoreService.GetParticipationScores() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParticipationScoreService.GetParticipationScores() = %v, want %v", got, tt.want)
			}
		})
	}
}
