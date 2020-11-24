package util

import (
	"reflect"
	"testing"
)

func TestSimpleQueueAddElement(t *testing.T) {
	type args struct {
		a []interface{}
		b interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  interface{}
		want1 []interface{}
	}{
		{
			name: "SimpleQueueAddElement",
			args: args{
				a: []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				b: 11,
			},
			want:  1,
			want1: []interface{}{2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := SimpleQueueAddElement(tt.args.a, tt.args.b)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimpleQueueAddElement() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SimpleQueueAddElement() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
