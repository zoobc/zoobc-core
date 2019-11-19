// Constants to generate/validate genesis block
// Note: this file has been auto-generated by 'genesis generate' command
package constant

const (
	MainchainGenesisBlockID int64 = -1701929749060110283
)

type (
	MainchainGenesisConfigEntry struct {
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
	MainChainGenesisConfig         = []MainchainGenesisConfigEntry{
		{
			AccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			AccountBalance: 0,
			NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118,
				97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			NodeAddress:        "172.104.34.10:8080",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress: "OnEYzI-EMV6UTfoUEzpQUjkSlnqB82-SyRN7469lJTWH",
			AccountBalance: 0,
			NodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152,
				194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255},
			NodeAddress:        "45.79.39.58:8080",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			AccountBalance: 0,
			NodePublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211, 123, 72, 52,
				221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
			NodeAddress:        "85.90.246.90:8080",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress: "nK_ouxdDDwuJiogiDAi_zs1LqeN7f5ZsXbFtXGqGc0Pd",
			AccountBalance: 100000000000,
			NodePublicKey: []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89, 234, 181, 100,
				163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220},
			NodeAddress:        "li1627-168.members.linode.com:8080",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress: "iSJt3H8wFOzlWKsy_UoEWF_OjF6oymHMqthyUMDKSyxb",
			AccountBalance: 100000000000,
			NodePublicKey: []byte{91, 36, 228, 70, 101, 94, 186, 246, 186, 4, 78, 142, 173, 162, 187, 173, 202, 81, 243,
				92, 141, 120, 148, 220, 41, 160, 208, 94, 174, 166, 62, 207},
			NodeAddress:        "172.104.47.168:8080",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
	}
)
