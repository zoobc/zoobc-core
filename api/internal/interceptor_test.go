package internal

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewInterceptor(t *testing.T) {

	type args struct {
		logger *logrus.Logger
	}
	tests := []struct {
		name string
		args args
		want grpc.UnaryServerInterceptor
	}{
		{
			name: "wantSuccess",
			args: args{
				logger: logrus.New(),
			},
			want: func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				return nil, status.Errorf(codes.Internal, "there's something wrong")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInterceptor(tt.args.logger); cmp.Equal(got, tt.want) {
				t.Errorf("NewInterceptor() = %v, want %v", got, tt.want)
			}
		})
	}
}
