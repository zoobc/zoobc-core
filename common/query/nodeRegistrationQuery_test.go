package query

import (
	"testing"
)

func TestGetNodeRegistrationNodeByPublicKey(t *testing.T) {

	nodeRegisterQuery := NewNodeRegistrationQuery()

	type paramsStruct struct {
		publicKey []byte
	}

	tests := []struct {
		name   string
		params *paramsStruct
		want   string
	}{
		{
			name: "Get node registration by public key",
			params: &paramsStruct{
				publicKey: []byte{4, 38, 113, 185},
			},
			want: "SELECT node_public_key, account_id, registration_height, node_address, " +
				"locked_balance, latest, height FROM node_registration " +
				"WHERE node_public_key = [4 38 113 185]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := nodeRegisterQuery.GetNodeRegistrationNodeByPublicKey(tt.params.publicKey)
			if query != tt.want {
				t.Errorf("GetNodeRegistrationNodeByPublicKey() \ngot = %v, \nwant = %v", query, tt.want)
				return
			}
		})
	}
}
