package handler

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	mockGetPendingTransactionsError struct {
		service.MultisigServiceInterface
	}
	mockGetPendingTransactionsSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetPendingTransactionsError) GetPendingTransactions(param *model.GetPendingTransactionsRequest,
) (*model.GetPendingTransactionsResponse, error) {
	return nil, errors.New("Error GetPendingTransactions")
}

func (*mockGetPendingTransactionsSuccess) GetPendingTransactions(param *model.GetPendingTransactionsRequest,
) (*model.GetPendingTransactionsResponse, error) {
	return &model.GetPendingTransactionsResponse{}, nil
}

func TestMultisigHandler_GetPendingTransactions(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetPendingTransactionsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionsResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactions:ErrorPageLessThanOne",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 0,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactions:Error",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetPendingTransactionsError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactions:Success",
			args: args{
				req: &model.GetPendingTransactionsRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetPendingTransactionsSuccess{},
			},
			want:    &model.GetPendingTransactionsResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetPendingTransactions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetPendingTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetPendingTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetPendingTransactionDetailByTransactionHashError struct {
		service.MultisigServiceInterface
	}
	mockGetPendingTransactionDetailByTransactionHashSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetPendingTransactionDetailByTransactionHashError) GetPendingTransactionDetailByTransactionHash(
	param *model.GetPendingTransactionDetailByTransactionHashRequest) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	return nil, errors.New("Error GetPendingTransactionDetailByTransactionHash")
}

func (*mockGetPendingTransactionDetailByTransactionHashSuccess) GetPendingTransactionDetailByTransactionHash(
	param *model.GetPendingTransactionDetailByTransactionHashRequest) (*model.GetPendingTransactionDetailByTransactionHashResponse, error) {
	return &model.GetPendingTransactionDetailByTransactionHashResponse{}, nil
}

func TestMultisigHandler_GetPendingTransactionDetailByTransactionHash(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetPendingTransactionDetailByTransactionHashRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetPendingTransactionDetailByTransactionHashResponse
		wantErr bool
	}{
		{
			name: "GetPendingTransactionDetailByTransactionHash:Error",
			fields: fields{
				MultisigService: &mockGetPendingTransactionDetailByTransactionHashError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPendingTransactionDetailByTransactionHash:Success",
			fields: fields{
				MultisigService: &mockGetPendingTransactionDetailByTransactionHashSuccess{},
			},
			want:    &model.GetPendingTransactionDetailByTransactionHashResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetPendingTransactionDetailByTransactionHash(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetPendingTransactionDetailByTransactionHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetPendingTransactionDetailByTransactionHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetMultisignatureInfoError struct {
		service.MultisigServiceInterface
	}
	mockGetMultisignatureInfoSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetMultisignatureInfoError) GetMultisignatureInfo(param *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	return nil, errors.New("Error GetMultisignatureInfo")
}

func (*mockGetMultisignatureInfoSuccess) GetMultisignatureInfo(param *model.GetMultisignatureInfoRequest,
) (*model.GetMultisignatureInfoResponse, error) {
	return &model.GetMultisignatureInfoResponse{}, nil
}

func TestMultisigHandler_GetMultisignatureInfo(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		ctx context.Context
		req *model.GetMultisignatureInfoRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMultisignatureInfoResponse
		wantErr bool
	}{
		{
			name: "GetMultisignatureInfo:ErrorPageLessThanOne",
			args: args{
				req: &model.GetMultisignatureInfoRequest{
					Pagination: &model.Pagination{
						Page: 0,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo:ErrorLimitMoreThan30",
			args: args{
				req: &model.GetMultisignatureInfoRequest{
					Pagination: &model.Pagination{
						Page: 31,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo:Error",
			args: args{
				req: &model.GetMultisignatureInfoRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetMultisignatureInfoError{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetMultisignatureInfo:Success",
			args: args{
				req: &model.GetMultisignatureInfoRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			fields: fields{
				MultisigService: &mockGetMultisignatureInfoSuccess{},
			},
			want:    &model.GetMultisignatureInfoResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetMultisignatureInfo(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetMultisignatureInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetMultisignatureInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockGetMultisigAddressByParticipantAddressesError struct {
		service.MultisigServiceInterface
	}
	mockGetMultisigAddressByParticipantAddressesSuccess struct {
		service.MultisigServiceInterface
	}
)

func (*mockGetMultisigAddressByParticipantAddressesError,
) GetMultisigAddressByParticipantAddresses(param *model.GetMultisigAddressByParticipantAddressesRequest,
) (*model.GetMultisigAddressByParticipantAddressesResponse, error) {
	return nil, errors.New("Error GetMultisigAddressByParticipantAddresses")
}

func (*mockGetMultisigAddressByParticipantAddressesSuccess,
) GetMultisigAddressByParticipantAddresses(param *model.GetMultisigAddressByParticipantAddressesRequest,
) (*model.GetMultisigAddressByParticipantAddressesResponse, error) {
	return &model.GetMultisigAddressByParticipantAddressesResponse{}, nil
}

func TestMultisigHandler_GetMultisigAddressByParticipantAddresses(t *testing.T) {
	type fields struct {
		MultisigService service.MultisigServiceInterface
	}
	type args struct {
		in0 context.Context
		req *model.GetMultisigAddressByParticipantAddressesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetMultisigAddressByParticipantAddressesResponse
		wantErr bool
	}{
		{
			name:   "getError:PageCannotBeLessThanOne",
			fields: fields{},
			args: args{
				req: &model.GetMultisigAddressByParticipantAddressesRequest{
					Pagination: &model.Pagination{
						Page: 0,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "getError:LimitCannotBeMoreThan30",
			fields: fields{},
			args: args{
				req: &model.GetMultisigAddressByParticipantAddressesRequest{
					Pagination: &model.Pagination{
						Page: 31,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "getError",
			fields: fields{
				MultisigService: &mockGetMultisigAddressByParticipantAddressesError{},
			},
			args: args{
				req: &model.GetMultisigAddressByParticipantAddressesRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "getSuccess",
			fields: fields{
				MultisigService: &mockGetMultisigAddressByParticipantAddressesSuccess{},
			},
			args: args{
				req: &model.GetMultisigAddressByParticipantAddressesRequest{
					Pagination: &model.Pagination{
						Page: 1,
					},
				},
			},
			want:    &model.GetMultisigAddressByParticipantAddressesResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msh := &MultisigHandler{
				MultisigService: tt.fields.MultisigService,
			}
			got, err := msh.GetMultisigAddressByParticipantAddresses(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultisigHandler.GetMultisigAddressByParticipantAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultisigHandler.GetMultisigAddressByParticipantAddresses() = %v, want %v", got, tt.want)
			}
		})
	}
}
