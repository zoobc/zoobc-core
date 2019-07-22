package query

import (
	"testing"
)

func TestGetTotalRecordOfSelect(t *testing.T) {
	tests := []struct {
		name        string
		selectQuery string
		want        string
	}{
		{
			name:        "Transforms record select query into count select query",
			selectQuery: "SELECT column1, column2 from any table",
			want:        "SELECT count() as total_record FROM any table",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := GetTotalRecordOfSelect(tt.selectQuery)
			if query != tt.want {
				t.Errorf("TestGetTotalRecordOfSelect() \ngot = %v, \nwant = %v", query, tt.want)
				return
			}
		})
	}
}
