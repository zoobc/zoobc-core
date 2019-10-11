package util

import "testing"

func TestRenderByteArrayAsString(t *testing.T) {
	type args struct {
		bArray []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestRenderByteArrayAsString:success",
			args: args{
				bArray: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118,
					97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			},
			want: `[]byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 
				97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RenderByteArrayAsString(tt.args.bArray); got != tt.want {
				t.Errorf("RenderByteArrayAsString() = %v, want %v", got, tt.want)
			}
		})
	}
}
