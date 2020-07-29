package service

import (
	"github.com/abiosoft/ishell"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"testing"
)

type mockConfigurationHelperService struct {
	NodeConfigurationServiceHelperInterface
	password string
}

func (nssHMock *mockConfigurationHelperService) ReadPassword(c *ishell.Shell) string {
	return nssHMock.password
}

func TestNodeConfigurationService_ImportWalletCertificate(t *testing.T) {
	type fields struct {
		Logger        *log.Logger
		host          *model.Host
		ServiceHelper NodeConfigurationServiceHelperInterface
	}
	type args struct {
		config *model.Config
	}
	tests := []struct {
		name                          string
		fields                        fields
		args                          args
		wantErr                       bool
		wantConfigNodeSeed            string
		wantConfigOwnerAccountAddress string
	}{
		{
			name: "importWalletCertificate:success",
			args: args{
				config: &model.Config{
					WalletCertFileName: "test-cert.zbc",
					ResourcePath:       "./testdata",
				},
			},
			fields: fields{
				Logger: log.New(),
				ServiceHelper: &mockConfigurationHelperService{
					password: "abcdefgh12345678",
				},
			},
			wantConfigOwnerAccountAddress: "ZBC_RERG3XD7_GAKOZZKY_VMZP2SQE_LBP45DC6_VDFGDTFK_3BZFBQGK_JMWELLO7",
			wantConfigNodeSeed:            "evolve list approve kangaroo fringe romance space kit idle shop major open",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nss := &NodeConfigurationService{
				Logger:        tt.fields.Logger,
				host:          tt.fields.host,
				ServiceHelper: tt.fields.ServiceHelper,
			}
			if err := nss.ImportWalletCertificate(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("ImportWalletCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.config.NodeSeed != tt.wantConfigNodeSeed || tt.args.config.OwnerAccountAddress != tt.wantConfigOwnerAccountAddress {
				t.Error("ImportWalletCertificate() decrypt error")
			}
		})
	}
}
