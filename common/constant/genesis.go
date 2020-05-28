// Constants to generate/validate genesis block
// Note: this file has been auto-generated by 'genesis generate' command
package constant

const (
	MainchainGenesisBlockID int64 = 7693829184545826152
)

type (
	GenesisConfigEntry struct {
		AccountAddress     string
		AccountBalance     int64
		NodePublicKey      []byte
		NodeAddress        string
		LockedBalance      int64
		ParticipationScore int64
	}
)

var (
	MainchainGenesisBlocksmithID   = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	MainchainGenesisBlockSignature = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	MainchainGenesisTransactionSignature = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	MainchainGenesisBlockTimestamp = int64(1562117271)
	MainchainGenesisAccountAddress = "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7"
	MainchainGenesisBlockSeed      = make([]byte, 64)
	MainchainGenesisNodePublicKey  = make([]byte, 32)
	// GenesisConfig one configuration for all block types (mainchain, spinechain), since they might share some fields
	GenesisConfig = []GenesisConfigEntry{
		{
			AccountAddress:     "2HoAsfL8ZnbkBOy2eGYqSbxzmOSk3lmN0mfh1J9J3XDe",
			AccountBalance:     0,
			NodePublicKey:      []byte{96, 21, 14, 132, 184, 49, 58, 139, 223, 5, 194, 185, 154, 93, 70, 25, 220, 39, 12, 90, 133, 239, 3, 248, 26, 144, 109, 252, 122, 153, 193, 107},
			NodeAddress:        "127.0.0.1:8001",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress:     "nIsUYYvfb5qxeY6kT5yU3Nxw5CjwOgSjSL53C4SNaE8_",
			AccountBalance:     0,
			NodePublicKey:      []byte{223, 12, 38, 9, 208, 50, 54, 70, 114, 245, 153, 140, 160, 228, 6, 40, 117, 43, 63, 89, 55, 101, 229, 192, 6, 100, 16, 43, 191, 232, 81, 98},
			NodeAddress:        "127.0.0.1:8002",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress:     "iSJt3H8wFOzlWKsy_UoEWF_OjF6oymHMqthyUMDKSyxb",
			AccountBalance:     100000000000,
			NodePublicKey:      []byte{230, 146, 58, 14, 220, 96, 98, 166, 87, 139, 81, 212, 206, 173, 44, 132, 235, 168, 253, 65, 79, 12, 193, 252, 46, 97, 167, 93, 65, 238, 147, 10},
			NodeAddress:        "127.0.0.1:8003",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
	}
)
