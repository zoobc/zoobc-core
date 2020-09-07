package decryptcert

type (
	clusterConfigEntry struct {
		NodePublicKey       string `json:"nodePublicKey"`
		NodeSeed            string `json:"nodeSeed"`
		OwnerAccountAddress string `json:"ownerAccountAddress"`
		NodeAddress         string `json:"myAddress,omitempty"`
		Smithing            bool   `json:"smithing,omitempty"`
	}
	certEntry struct {
		NodeSeed            string `json:"nodeKey"`
		OwnerAccountAddress string `json:"ownerAccount"`
	}
	encryptedCertEntry struct {
		EncryptedCert string `json:"encryptedCert"`
		Password      string `json:"password"`
	}
)
