package util

import (
	"fmt"
	"reflect"
	"time"
)

type (
	Scheduler struct {
		Wait *time.Ticker
	}
)

// NewScheduler return new scheduler instance
func NewScheduler(period time.Duration) *Scheduler {
	return &Scheduler{
		Wait: time.NewTicker(period),
	}
}

// AddJob task runner repeatably in some period time
func (s *Scheduler) AddJob(fn interface{}, args ...interface{}) error {

	f := reflect.ValueOf(fn)
	if f.Kind() != reflect.Func {
		return fmt.Errorf("the fn is not function")
	}
	if len(args) != f.Type().NumIn() {
		return fmt.Errorf("the argument of function not match, %d needed", f.Type().NumIn())
	}

	params := make([]reflect.Value, len(args))
	for k, arg := range args {
		params[k] = reflect.ValueOf(arg)
	}

	for {
		<-s.Wait.C
		f.Call(params)
	}
}

// Stop handle to stop scheduler
func (s *Scheduler) Stop() {
	s.Wait.Stop()
}
