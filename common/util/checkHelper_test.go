package util

import "testing"

func TestValidateIP4(t *testing.T) {
	type args struct {
		ipAddress string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "TestValidateIP4:true",
			args: args{
				ipAddress: "127.0.0.1",
			},
			want: true,
		},
		{
			name: "TestValidateIP4:false",
			args: args{
				ipAddress: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateIP4(tt.args.ipAddress); got != tt.want {
				t.Errorf("ValidateIP4() = %v, want %v", got, tt.want)
			}
		})
	}
}
