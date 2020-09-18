package decryptcert

type (
	clusterConfigEntry struct {
		NodePublicKey  string `json:"NodePublicKey"`
		NodeSeed       string `json:"NodeSeed"`
		AccountAddress string `json:"AccountAddress"`
		NodeAddress    string `json:"MyAddress,omitempty"`
		Smithing       bool   `json:"Smithing,omitempty"`
	}
	certEntry struct {
		NodeSeed       string `json:"nodeSeed"`
		AccountAddress string `json:"ownerAccount"`
		NodePublicKey  string `json:"nodePublicKey"`
	}
	encryptedCertEntry struct {
		EncryptedCert string `json:"encryptedCert"`
		Password      string `json:"password"`
	}
)
