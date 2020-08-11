package blocker

import (
	"fmt"
	"testing"
)

func TestBlocker_Error(t *testing.T) {
	type fields struct {
		Type    TypeBlocker
		Message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Error",
			fields: fields{
				Type:    DBErr,
				Message: "sql error",
			},
			want: "DBErr: sql error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Blocker{
				Type:    tt.fields.Type,
				Message: tt.fields.Message,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Blocker.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlocker(t *testing.T) {
	type args struct {
		typeBlocker TypeBlocker
		message     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "WantError",
			args: args{
				typeBlocker: BlockErr,
				message:     "invalid block height",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewBlocker(tt.args.typeBlocker, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("NewBlocker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ExampleNewBlocker() {
	err := NewBlocker(BlockErr, "invalid block height")
	fmt.Println(err)
	// Output: BlockErr: invalid block height
}
