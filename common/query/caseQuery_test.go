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
		want   *CaseQuery
	}{
		{
			name: "SelectFields",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
			},
			args: args{
				tableName: "account",
				columns:   []string{"id", "name"},
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("SELECT id, name FROM account "),
			},
		},
		{
			name: "SelectAll",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
			},
			args: args{
				tableName: "account",
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("SELECT * FROM account "),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			if got := fq.Select(tt.args.tableName, tt.args.columns...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Select() = %v, want %v", got, tt.want)
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
		query []string
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1, "bcz")

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CaseQuery
	}{
		{
			name: "Where",
			fields: fields{
				Query: bytes.NewBufferString("SELECT id, name FROM account "),
				Args:  argsWant,
			},
			args: args{
				query: []string{
					"id = ? ",
					"AND name = ? ",
				},
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("SELECT id, name FROM account WHERE id = ? AND name = ?  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			if got := fq.Where(tt.args.query...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Where() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCaseQuery_And(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		expression []string
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1, "bcz")

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CaseQuery
	}{
		{
			name: "And",
			fields: fields{
				Query: bytes.NewBufferString("SELECT id, name FROM account WHERE id = ? "),
				Args:  argsWant,
			},
			args: args{
				expression: []string{
					"name = ? ",
				},
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("SELECT id, name FROM account WHERE id = ? AND name = ? "),
				Args:  argsWant,
			},
		},
		{
			name: "AndWithoutWHERE",
			fields: fields{
				Query: bytes.NewBufferString("SELECT id, name FROM account "),
				Args:  argsWant,
			},
			args: args{
				expression: []string{
					"name = ? ",
				},
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("SELECT id, name FROM account WHERE 1=1 AND name = ? "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			if got := fq.And(tt.args.expression...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("And() = %v, want %v", got, tt.want)
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
		expression []string
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1, "bcz")

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CaseQuery
	}{
		{
			name: "Or",
			fields: fields{
				Query: bytes.NewBufferString("SELECT id, name FROM account WHERE id = ? "),
				Args:  argsWant,
			},
			args: args{
				expression: []string{
					"name = ?",
				},
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("SELECT id, name FROM account WHERE id = ? OR name = ? "),
				Args:  argsWant,
			},
		},
		{
			name: "OrWithoutWHERE",
			fields: fields{
				Query: bytes.NewBufferString("SELECT id, name FROM account "),
				Args:  argsWant,
			},
			args: args{
				expression: []string{
					"name = ?",
				},
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("SELECT id, name FROM account WHERE 1=1 OR name = ? "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			if got := fq.Or(tt.args.expression...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CaseQuery.Or() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCaseQuery_In(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		column string
		value  []interface{}
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1, 2)

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          string
		wantCaseQuery *CaseQuery
	}{
		{
			name: "In",
			fields: fields{
				Query: bytes.NewBufferString(""),
			},
			args: args{
				column: "id",
				value:  []interface{}{1, 2},
			},
			want: "id IN (?, ?) ",
			wantCaseQuery: &CaseQuery{
				Query: bytes.NewBufferString("WHERE id IN (?, ?)  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}

			got := fq.In(tt.args.column, tt.args.value...)
			if got != tt.want {
				t.Errorf("CaseQuery.In() = %v, want %v", got, tt.want)
				return
			}

			fq.Where(got) // represents fq.Where(fq.In())
			if !reflect.DeepEqual(tt.wantCaseQuery, fq) {
				t.Errorf("CaseQuery.In() = %v want %v", fq, tt.wantCaseQuery)
			}
		})
	}
}

func TestCaseQuery_NotIn(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		column string
		value  []interface{}
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1, 2)

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          string
		wantCaseQuery *CaseQuery
	}{
		{
			name: "In",
			fields: fields{
				Query: bytes.NewBufferString(""),
			},
			args: args{
				column: "id",
				value:  []interface{}{1, 2},
			},
			want: "id NOT IN (?, ?) ",
			wantCaseQuery: &CaseQuery{
				Query: bytes.NewBufferString("WHERE id NOT IN (?, ?)  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			got := fq.NotIn(tt.args.column, tt.args.value...)
			if got != tt.want {
				t.Errorf("CaseQuery.NotIn() = %v, want %v", got, tt.want)
				return
			}
			fq.Where(got) // represents fq.Where(fq.NotIn())
			if !reflect.DeepEqual(tt.wantCaseQuery, fq) {
				t.Errorf("CaseQuery.NotIn() = %v want %v", fq, tt.wantCaseQuery)
			}
		})
	}
}

func TestCaseQuery_Equal(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		column string
		value  interface{}
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1)

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          string
		wantCaseQuery *CaseQuery
	}{
		{
			name: "Equal",
			fields: fields{
				Query: bytes.NewBufferString(""),
			},
			args: args{
				column: "id",
				value:  1,
			},
			want: "id = ? ",
			wantCaseQuery: &CaseQuery{
				Query: bytes.NewBufferString("WHERE id = ?  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			got := fq.Equal(tt.args.column, tt.args.value)
			if got != tt.want {
				t.Errorf("CaseQuery.Equal() = %v, want %v", got, tt.want)
				return
			}
			fq.Where(got) // represents fq.Where(fq.Equal())
			if !reflect.DeepEqual(fq, tt.wantCaseQuery) {
				t.Errorf("CaseQuery.Equal() = %v, want %v", fq, tt.wantCaseQuery)
			}
		})
	}
}

func TestCaseQuery_NotEqual(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		column string
		value  interface{}
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1)

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          string
		wantCaseQuery *CaseQuery
	}{
		{
			name: "Equal",
			fields: fields{
				Query: bytes.NewBufferString(""),
			},
			args: args{
				column: "id",
				value:  1,
			},
			want: "id <> ? ",
			wantCaseQuery: &CaseQuery{
				Query: bytes.NewBufferString("WHERE id <> ?  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			got := fq.NotEqual(tt.args.column, tt.args.value)
			if got != tt.want {
				t.Errorf("CaseQuery.NotEqual() = %v, want %v", got, tt.want)
				return
			}
			fq.Where(got)
			if !reflect.DeepEqual(fq, tt.wantCaseQuery) {
				t.Errorf("CaseQuery.NotEqual() = %v, want %v", fq, tt.wantCaseQuery)
			}
		})
	}
}

func TestCaseQuery_GreaterEqual(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		column string
		value  interface{}
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 2)

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          string
		wantCaseQuery *CaseQuery
	}{
		{
			name: "GreaterEqual",
			fields: fields{
				Query: bytes.NewBufferString(""),
			},
			args: args{
				column: "id",
				value:  2,
			},
			want: "id >= ? ",
			wantCaseQuery: &CaseQuery{
				Query: bytes.NewBufferString("WHERE id >= ?  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			got := fq.GreaterEqual(tt.args.column, tt.args.value)
			if got != tt.want {
				t.Errorf("CaseQuery.GreaterEqual() = %v, want %v", got, tt.want)
			}
			fq.Where(got)
			if !reflect.DeepEqual(fq, tt.wantCaseQuery) {
				t.Errorf("CaseQuery.NotEqual() = %v, want %v", fq, tt.wantCaseQuery)
			}
		})
	}
}

func TestCaseQuery_LessEqual(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		column string
		value  interface{}
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 2)

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          string
		wantCaseQuery *CaseQuery
	}{
		{
			name: "LessEqual",
			fields: fields{
				Query: bytes.NewBufferString(""),
			},
			args: args{
				column: "id",
				value:  2,
			},
			want: "id <= ? ",
			wantCaseQuery: &CaseQuery{
				Query: bytes.NewBufferString("WHERE id <= ?  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			got := fq.LessEqual(tt.args.column, tt.args.value)
			if got != tt.want {
				t.Errorf("CaseQuery.LessEqual() = %v, want %v", got, tt.want)
			}

			fq.Where(got)
			if !reflect.DeepEqual(fq, tt.wantCaseQuery) {
				t.Errorf("CaseQuery.NotEqual() = %v, want %v", fq, tt.wantCaseQuery)
			}
		})
	}
}

func TestCaseQuery_Between(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		column string
		start  interface{}
		end    interface{}
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1000, 1234)

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          string
		wantCaseQuery *CaseQuery
	}{
		{
			name: "Between",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
			},
			args: args{
				column: "timestamp",
				start:  1000,
				end:    1234,
			},
			want: "timestamp BETWEEN ? AND ? ",
			wantCaseQuery: &CaseQuery{
				Query: bytes.NewBufferString("WHERE timestamp BETWEEN ? AND ?  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			got := fq.Between(tt.args.column, tt.args.start, tt.args.end)
			if got != tt.want {
				t.Errorf("CaseQuery.Between() = %v, want %v", got, tt.want)
				return
			}
			fq.Where(got)
			if !reflect.DeepEqual(fq, tt.wantCaseQuery) {
				t.Errorf("CaseQuery.Between() = %v, want %v", fq, tt.wantCaseQuery)
			}
		})
	}
}

func TestCaseQuery_NotBetween(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		column string
		start  interface{}
		end    interface{}
	}
	var argsWant []interface{}
	argsWant = append(argsWant, 1000, 1234)

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          string
		wantCaseQuery *CaseQuery
	}{
		{
			name: "NotBetween",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
			},
			args: args{
				column: "timestamp",
				start:  1000,
				end:    1234,
			},
			want: "timestamp NOT BETWEEN ? AND ? ",
			wantCaseQuery: &CaseQuery{
				Query: bytes.NewBufferString("WHERE timestamp NOT BETWEEN ? AND ?  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			got := fq.NotBetween(tt.args.column, tt.args.start, tt.args.end)
			if got != tt.want {
				t.Errorf("CaseQuery.NotBetween() = %v, want %v", got, tt.want)
				return
			}
			fq.Where(got)
			if !reflect.DeepEqual(fq, tt.wantCaseQuery) {
				t.Errorf("CaseQuery.NotBetween() = %v, want %v", fq, tt.wantCaseQuery)
			}
		})
	}
}

func TestCaseQuery_Paginate(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	type args struct {
		limit       uint32
		currentPage uint32
	}
	var argsZeroWant, argsWant []interface{}
	argsZeroWant = append(argsZeroWant, uint32(30), uint32(0))
	argsWant = append(argsWant, uint32(1), uint32(0))

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CaseQuery
	}{
		{
			name: "PaginateZero",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
			},
			args: args{
				limit:       0,
				currentPage: 0,
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("limit ? offset ?  "),
				Args:  argsZeroWant,
			},
		},
		{
			name: "Paginate",
			fields: fields{
				Query: bytes.NewBuffer([]byte{}),
			},
			args: args{
				limit:       1,
				currentPage: 1,
			},
			want: &CaseQuery{
				Query: bytes.NewBufferString("limit ? offset ?  "),
				Args:  argsWant,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}

			if got := fq.Paginate(tt.args.limit, tt.args.currentPage); reflect.DeepEqual(got, tt.want) {
				t.Errorf("CaseQuery.Paginate() = \n%v, want \n%v", fq, tt.want)
			}
		})
	}
}

func TestCaseQuery_Build(t *testing.T) {
	type fields struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
	tests := []struct {
		name      string
		fields    fields
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "Build",
			fields: fields{
				Query: bytes.NewBufferString("SELECT id, name FROM account"),
				Args:  make([]interface{}, 0),
			},
			wantQuery: "SELECT id, name FROM account",
			wantArgs:  make([]interface{}, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &CaseQuery{
				Query: tt.fields.Query,
				Args:  tt.fields.Args,
			}
			gotQuery, gotArgs := fq.Build()
			if gotQuery != tt.wantQuery {
				t.Errorf("CaseQuery.Build() gotQuery = %v, want %v", gotQuery, tt.wantQuery)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("CaseQuery.Build() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
