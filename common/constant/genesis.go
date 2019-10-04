// Constants to generate/validate genesis block
package constant

const (
	MainchainGenesisBlockID int64 = -1294179708803500770
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
	MainChainGenesisConfig         = []struct {
		AccountAddress     string
		AccountBalance     int64
		NodePublicKey      []byte
		NodeAddress        string
		LockedBalance      int64
		ParticipationScore int64
	}{
		{
			AccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			AccountBalance: 0,
			NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99,
				125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			NodeAddress:        "0.0.0.0",
			LockedBalance:      100000 * OneZBC,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			AccountBalance: 0,
			NodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126,
				203, 5, 12, 152, 194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255},
			NodeAddress:        "0.0.0.0",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			AccountBalance: 0,
			NodePublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211, 123,
				72, 52, 221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
			NodeAddress:        "0.0.0.0",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
		{
			AccountAddress: "nK_ouxdDDwuJiogiDAi_zs1LqeN7f5ZsXbFtXGqGc0Pd",
			AccountBalance: 1000 * OneZBC,
			NodePublicKey: []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89,
				234, 181, 100, 163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220},
			NodeAddress:        "0.0.0.0",
			LockedBalance:      0,
			ParticipationScore: DefaultParticipationScore,
		},
	}
)
