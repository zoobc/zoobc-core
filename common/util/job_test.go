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
		schedulerInterval time.Duration
	}
	tests := []struct {
		name string
		args args
		want *Scheduler
	}{
		{
			name: "wanScheduler",
			args: args{
				schedulerInterval: time.Millisecond * 30,
			},
			want: &Scheduler{
				Jobs:     make(map[string]*Job),
				Interval: time.Millisecond * 30,
				Done:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScheduler(tt.args.schedulerInterval); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScheduler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduler_AddJob(t *testing.T) {
	type fields struct {
		Jobs     map[string]*Job
		Interval time.Duration
		Done     bool
	}
	type args struct {
		jobName string
		period  time.Duration
		fn      interface{}
		args    []interface{}
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
				jobName: "jobSuccess",
				period:  time.Millisecond,
				fn:      func(str string) { fmt.Println(str) },
				args:    []interface{}{"Ariasa"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scheduler{
				Jobs:     tt.fields.Jobs,
				Interval: tt.fields.Interval,
				Done:     tt.fields.Done,
			}
			if err := s.AddJob(tt.args.jobName, tt.args.period, tt.args.fn, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Scheduler.AddJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
