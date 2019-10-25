package util

import (
	"fmt"
	"reflect"
	"time"
)

type (
	Scheduler struct {
		Done         chan bool
		NumberOfJobs int
	}
)

// NewScheduler return new scheduler instance
func NewScheduler() *Scheduler {
	return &Scheduler{
		Done: make(chan bool),
	}
}

// AddJob adding new job in scheduler
func (s *Scheduler) AddJob(period time.Duration, fn interface{}, args ...interface{}) error {
	var (
		jobFunction = reflect.ValueOf(fn)
		jobParams   = make([]reflect.Value, len(args))
	)
	if jobFunction.Kind() != reflect.Func {
		return fmt.Errorf("the fn is not function")
	}
	if len(args) != jobFunction.Type().NumIn() {
		return fmt.Errorf("the argument of function not match, %d needed", jobFunction.Type().NumIn())
	}
	for k, arg := range args {
		jobParams[k] = reflect.ValueOf(arg)
	}

	go func() {
		for {
			select {
			case <-s.Done:
				return
			case <-time.NewTicker(period).C:
				jobFunction.Call(jobParams)
			}
		}
	}()
	s.NumberOfJobs = s.NumberOfJobs + 1
	return nil
}

// Stop handle to stop scheduler
func (s *Scheduler) Stop() {
	for stopped := 0; stopped < s.NumberOfJobs; stopped++ {
		s.Done <- true
	}
}
