package feedbacksystem

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewDummyFeedbackStrategy(t *testing.T) {
	tests := []struct {
		name string
		want *DummyFeedbackStrategy
	}{
		{
			name: "NewAntiSpamStrategy:success",
			want: &DummyFeedbackStrategy{
				FeedbackVars: map[string]interface{}{
					"tpsReceived":         0,
					"tpsReceivedTmp":      0,
					"tpsProcessed":        0,
					"tpsProcessedTmp":     0,
					"txReceived":          0,
					"txProcessed":         0,
					"P2PIncomingRequests": 0,
					"P2POutgoingRequests": 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDummyFeedbackStrategy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDummyFeedbackStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyFeedbackStrategy_GetFeedbackVar(t *testing.T) {
	type fields struct {
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
	}
	type args struct {
		k string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "GetFeedbackVar",
			fields: fields{
				FeedbackVars: map[string]interface{}{
					"testVar": 10,
				},
			},
			args: args{
				k: "testVar",
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dfs := &DummyFeedbackStrategy{
				FeedbackVars:     tt.fields.FeedbackVars,
				FeedbackVarsLock: tt.fields.FeedbackVarsLock,
			}
			if got := dfs.GetFeedbackVar(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFeedbackVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyFeedbackStrategy_IncrementVarCount(t *testing.T) {
	type fields struct {
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
	}
	type args struct {
		k string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "IncrementVarCount:success",
			fields: fields{
				FeedbackVars: map[string]interface{}{
					"testVar": 10,
				},
			},
			args: args{
				k: "testVar",
			},
			want: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dfs := &DummyFeedbackStrategy{
				FeedbackVars:     tt.fields.FeedbackVars,
				FeedbackVarsLock: tt.fields.FeedbackVarsLock,
			}
			if got := dfs.IncrementVarCount(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IncrementVarCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyFeedbackStrategy_DecrementVarCount(t *testing.T) {
	type fields struct {
		FeedbackVars     map[string]interface{}
		FeedbackVarsLock sync.RWMutex
	}
	type args struct {
		k string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "DecrementVarCount:success",
			fields: fields{
				FeedbackVars: map[string]interface{}{
					"testVar": 10,
				},
			},
			args: args{
				k: "testVar",
			},
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dfs := &DummyFeedbackStrategy{
				FeedbackVars:     tt.fields.FeedbackVars,
				FeedbackVarsLock: tt.fields.FeedbackVarsLock,
			}
			if got := dfs.DecrementVarCount(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecrementVarCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
