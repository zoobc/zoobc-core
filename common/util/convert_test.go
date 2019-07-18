package util

import (
	"reflect"
	"testing"
)

func TestConvertBytesToUint64(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{

		{
			name: "ConvertBytesToUint64:one",
			args: args{
				[]byte{12, 43, 54, 45, 12, 5, 2, 5},
			},
			want: 360856469999332108,
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				[]byte{12, 43, 54, 45, 12, 5, 2, 54},
			},
			want: 3891678577857800972,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesToUint64(tt.args.bytes); got != tt.want {
				t.Errorf("ConvertBytesToUint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertBytesToUint32(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				[]byte{12, 43, 54, 45},
			},
			want: 758524684,
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				[]byte{54, 23, 54, 45},
			},
			want: 758519606,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesToUint32(tt.args.bytes); got != tt.want {
				t.Errorf("ConvertBytesToUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertBytesToUint16(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				[]byte{12, 43},
			},
			want: 11020,
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				[]byte{54, 23},
			},
			want: 5942,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesToUint16(tt.args.bytes); got != tt.want {
				t.Errorf("ConvertBytesToUint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertUint64ToBytes(t *testing.T) {
	type args struct {
		number uint64
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				360856469999332108,
			},
			want: []byte{12, 43, 54, 45, 12, 5, 2, 5},
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				3891678577857800972,
			},
			want: []byte{12, 43, 54, 45, 12, 5, 2, 54},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertUint64ToBytes(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertUint64ToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertUint32ToBytes(t *testing.T) {
	type args struct {
		number uint32
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				758524684,
			},
			want: []byte{12, 43, 54, 45},
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				758519606,
			},
			want: []byte{54, 23, 54, 45},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertUint32ToBytes(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertUint32ToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertUint16ToBytes(t *testing.T) {
	type args struct {
		number uint16
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				11020,
			},
			want: []byte{12, 43},
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				5942,
			},
			want: []byte{54, 23},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertUint16ToBytes(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertUint16ToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertIntToBytes(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertIntToByte:one",
			args: args{
				5942,
			},
			want: []byte{54, 23, 0, 0},
		},
		{
			name: "ConvertIntToByte:two",
			args: args{
				11020,
			},
			want: []byte{12, 43, 0, 0},
		},
		{
			name: "ConvertIntToByte:three",
			args: args{
				758519606,
			},
			want: []byte{54, 23, 54, 45},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertIntToBytes(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertIntToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertStringToBytes(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertStringToBytes:success",
			args: args{
				str: "dummy random string here",
			},
			want: []byte{24, 0, 0, 0, 100, 117, 109, 109, 121, 32, 114, 97, 110, 100, 111, 109, 32, 115, 116, 114, 105, 110, 103, 32,
				104, 101, 114, 101},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertStringToBytes(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertStringToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
