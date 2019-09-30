package util

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

var (
	mockScheduler = NewScheduler(30 * time.Millisecond)
)

func TestNewScheduler(t *testing.T) {
	type args struct {
		period time.Duration
	}
	duration := time.Duration(3000)
	tests := []struct {
		name string
		args args
		want *Scheduler
	}{
		{
			name: "wanScheduler",
			args: args{period: duration},
			want: &Scheduler{Wait: time.NewTicker(duration)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScheduler(tt.args.period); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("NewScheduler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduler_AddJob(t *testing.T) {
	type fields struct {
		Wait *time.Ticker
	}
	type args struct {
		fn   interface{}
		args []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Error:IsNotFunc",
			fields: fields(*mockScheduler),
			args: args{
				fn:   "",
				args: []interface{}{"Test"},
			},
			wantErr: true,
		},
		{
			name:   "Error:NotMatchArgs",
			fields: fields(*mockScheduler),
			args: args{
				fn:   func() {},
				args: []interface{}{"asdsdsd"},
			},
			wantErr: true,
		},
		{
			name:   "Success",
			fields: fields(*mockScheduler),
			args: args{
				fn:   func(str string) { fmt.Println(str) },
				args: []interface{}{"Ariasa"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scheduler{
				Wait: tt.fields.Wait,
			}

			go func() {
				if err := s.AddJob(tt.args.fn, tt.args.args...); (err != nil) != tt.wantErr {
					t.Errorf("AddJob() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}()
			time.Sleep(2 * time.Millisecond)
			s.Stop()
		})
	}
}
