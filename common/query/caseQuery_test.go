package query

import (
	"bytes"
	"reflect"
	"testing"
)

func TestCaseQuery_Select(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		tableName string
		columns   []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *fields
	}{
		{
			name: "Select",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
				Args:  make([]interface{}, 0),
			},
			args: args{
				tableName: "blocks",
				columns: []string{
					"id",
				},
			},
			want: &fields{
				Query: bytes.NewBufferString("SELECT id FROM blocks "), // make sure has last space,
				Args:  make([]interface{}, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			fq.Select(tt.args.tableName, tt.args.columns...)
			if !reflect.DeepEqual(tt.want.Query, fq.Query) || !reflect.DeepEqual(tt.want.Args, fq.Args) {
				t.Errorf("Select() got = \n%s, want = \n%s", fq, tt.want)
			}
		})
	}
}

func TestCaseQuery_Where(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		args map[string]interface{}
	}
	var (
		wantArgs []interface{}
	)
	wantArgs = append(wantArgs, 1, "bcz")

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *fields
	}{
		{
			name: "Where",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
				Args:  make([]interface{}, 0),
			},
			args: args{
				args: map[string]interface{}{
					"id":   1,
					"name": "bcz",
				},
			},
			want: &fields{
				Query: bytes.NewBufferString("WHERE id = ? AND name = ? "), // make sure space the last,
				Args:  wantArgs,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			fq.Where(tt.args.args)
			if !reflect.DeepEqual(fq.Query, tt.want.Query) || !reflect.DeepEqual(tt.want.Args, fq.Args) {
				t.Errorf("Where() want = %s, got = %s", tt.want, fq)
			}
		})
	}
}

func TestCaseQuery_Or(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		args map[string]interface{}
	}
	var (
		wantArgs []interface{}
	)
	wantArgs = append(wantArgs, 1, "bcz")

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *fields
	}{
		{
			name: "Or",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
				Args:  make([]interface{}, 0),
			},
			args: args{
				args: map[string]interface{}{
					"id":   1,
					"name": "bcz",
				},
			},
			want: &fields{
				Query: bytes.NewBufferString("OR id = ? AND name = ? "), // make sure space the last,
				Args:  wantArgs,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			fq.Or(tt.args.args)
			if (reflect.DeepEqual(fq.Query, tt.want.Query)) || (fq.Query.String() == "OR name = ? AND id = ? ") {
			} else {
				t.Errorf("Or() fq.Query is not equal with want.Query. got = \n%v, want = \n%v,", fq.Query, tt.want.Query)
				return
			}
			if !reflect.DeepEqual(tt.want.Args, fq.Args) {
				t.Errorf("Or() want = %s, got = %s", tt.want, fq)
			}

		})
	}
}

func TestCaseQuery_Conjunct(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		firstSep Operand
		args     map[string]interface{}
	}
	var (
		wantArgs []interface{}
	)
	wantArgs = append(wantArgs, 1, "bcz")

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *fields
	}{
		{
			name: "Conjuct",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
				Args:  make([]interface{}, 0),
			},
			args: args{
				firstSep: OperandOr,
				args: map[string]interface{}{
					"id":   1,
					"name": "bcz",
				},
			},
			want: &fields{
				Query: bytes.NewBufferString(" OR id = ? AND name = ? "), // make sure space in the last
				Args:  wantArgs,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			fq.Conjunct(tt.args.firstSep, tt.args.args)
			if (reflect.DeepEqual(fq.Query, tt.want.Query)) || (fq.Query.String() == "OR name = ? AND id = ? ") {
			} else {
				t.Errorf("Or() fq.Query is not equal with want.Query. got = \n%v, want = \n%v,", fq.Query, tt.want.Query)
				return
			}
			if !reflect.DeepEqual(tt.want.Args, fq.Args) {
				t.Errorf("Where() want = %s, got = %s", tt.want, fq)
			}
		})
	}
}

func TestCaseQuery_Done(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		limit  uint32
		offset uint64
	}
	var (
		values []interface{}
	)
	values = append(values, 1, 1)

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "Done",
			fields: fields{
				Query: bytes.NewBufferString("SELECT id, name FROM blocks "),
				Args:  make([]interface{}, 0),
			},
			args: args{
				limit:  1,
				offset: 1,
			},
			wantQuery: "SELECT id, name FROM blocks limit ? offset ? ",
			wantArgs:  values,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			gotQuery, gotArgs := fq.Done(tt.args.limit, tt.args.offset)
			if gotQuery != tt.wantQuery {
				t.Errorf("CaseQuery.Done() gotQuery = %v, want %v", gotQuery, tt.wantQuery)
				return
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("CaseQuery.Done() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
