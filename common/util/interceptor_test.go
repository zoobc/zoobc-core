package util

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewServerInterceptor(t *testing.T) {

	type args struct {
		logger *logrus.Logger
	}
	tests := []struct {
		name        string
		args        args
		want        grpc.UnaryServerInterceptor
		wantRecover bool
	}{
		{
			name: "wantRecover",
			args: args{
				logger: logrus.New(),
			},
			want: func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				return nil, status.Errorf(codes.Internal, "there's something wrong")
			},
			wantRecover: true,
		},
		{
			name: "wantNotRecover",
			args: args{
				logger: logrus.New(),
			},
			want: func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				return nil, status.Errorf(codes.Internal, "there's something wrong")
			},
			wantRecover: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServerInterceptor(tt.args.logger)
			if cmp.Equal(got, tt.want) {
				t.Errorf("NewInterceptor() = %v, want %v", got, tt.want)
			}
			testInterceptor(got, tt.wantRecover)
		})
	}
}

func testInterceptor(fn grpc.UnaryServerInterceptor, wantRecover bool) {
	var (
		handler grpc.UnaryHandler
	)
	if wantRecover {
		handler = func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			panic(handler)
		}
	} else {
		handler = func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			return nil, status.Errorf(codes.Internal, "there's something wrong")
		}
	}
	_, _ = fn(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
}
