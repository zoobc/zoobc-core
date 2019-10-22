package util

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type (
	Job struct {
		WaitDuration time.Duration
		NextSchedule time.Duration
		Function     reflect.Value
		Params       []reflect.Value
	}
	Scheduler struct {
		Jobs     map[string]*Job
		Interval time.Duration
		Done     bool
	}
)

// NewScheduler return new scheduler instance
func NewScheduler(schedulerInterval time.Duration) *Scheduler {
	return &Scheduler{
		Jobs:     make(map[string]*Job),
		Interval: schedulerInterval,
		Done:     false,
	}
}

// AddJob adding new job in scheduler
func (s *Scheduler) AddJob(jobName string, period time.Duration, fn interface{}, args ...interface{}) error {
	var (
		jobFunc   = reflect.ValueOf(fn)
		jobParams = make([]reflect.Value, len(args))
	)
	if jobFunc.Kind() != reflect.Func {
		return fmt.Errorf("the fn is not function")
	}
	if len(args) != jobFunc.Type().NumIn() {
		return fmt.Errorf("the argument of function not match, %d needed", jobFunc.Type().NumIn())
	}
	for k, arg := range args {
		jobParams[k] = reflect.ValueOf(arg)
	}
	job := Job{
		WaitDuration: period,
		NextSchedule: time.Duration(time.Now().Unix())*time.Second + period,
		Function:     jobFunc,
		Params:       jobParams,
	}
	s.Jobs[jobName] = &job
	return nil
}

// Start Scheduler running all job repeatably in some period time
func (s *Scheduler) Start() error {
	for {
		for _, job := range s.Jobs {
			if job.NextSchedule <= time.Duration(time.Now().Unix())*time.Second {
				job.NextSchedule = time.Duration(time.Now().Unix())*time.Second + job.WaitDuration
				job.Function.Call(job.Params)
			}
			time.Sleep(s.Interval)
		}
		if s.Done {
			return errors.New("Scheduler stoping all job")
		}
	}
}

// Stop handle to stop scheduler
func (s *Scheduler) Stop() {
	s.Done = true
}
