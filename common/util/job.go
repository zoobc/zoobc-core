package util

import (
	"fmt"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	Scheduler struct {
		Done         chan bool
		NumberOfJobs int
		Logger       *logrus.Logger
	}
)

// NewScheduler return new scheduler instance
func NewScheduler(logger *logrus.Logger) *Scheduler {
	return &Scheduler{
		Done:   make(chan bool),
		Logger: logger,
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
		ticker := time.NewTicker(period)
		for {
			select {
			case <-s.Done:
				ticker.Stop()
				return
			case <-ticker.C:
				// Execute method and log the error value
				values := jobFunction.Call(jobParams)
				if len(values) > 0 && !values[0].IsNil() {
					rf := reflect.ValueOf(values[0]).Interface()
					func() {
						s.Logger.Error(rf)
					}()
				}
			}
		}
	}()
	s.NumberOfJobs++
	return nil
}

// Stop handle to stop scheduler
func (s *Scheduler) Stop() {
	for stopped := 0; stopped < s.NumberOfJobs; stopped++ {
		s.Done <- true
	}
}
