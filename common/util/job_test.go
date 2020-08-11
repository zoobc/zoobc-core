package util

import (
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	mockScheduler = NewScheduler(logrus.New())
)

func TestScheduler_AddJob(t *testing.T) {
	type fields struct {
		Done         chan bool
		NumberOfJobs int
		Logger       *logrus.Logger
	}
	type args struct {
		period time.Duration
		fn     interface{}
		args   []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Success",
			fields: fields(*mockScheduler),
			args: args{
				period: time.Millisecond,
				fn: func(str string) error {
					return errors.New("need error")
				},
				args: []interface{}{"Test"},
			},
			wantErr: false,
		},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scheduler{
				Done:         tt.fields.Done,
				NumberOfJobs: tt.fields.NumberOfJobs,
				Logger:       tt.fields.Logger,
			}
			if err := s.AddJob(tt.args.period, tt.args.fn, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("Scheduler.AddJob() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(time.Millisecond + 10)
			s.Stop()
		})
	}
}
