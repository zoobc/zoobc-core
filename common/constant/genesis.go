// Constants to generate/validate genesis block
// TODO: in future this file could be automatically generated by a script
package constant

const (
	GenesisBlocID int64 = -5061068901394437496
)

var (
	// GenesisFundReceiver stake holders account data.
	// Note: 1 ZOO = 100000000 ZOOBIT, node only know the zoobit representation, zoo representation is handled by frontend
	GenesisFundReceiver = []struct {
		AccountAddress string
		Amount         int64
		NodePublicKey  []byte
		NodeAddress    string
		LockedBalance  int64
	}{
		{
			// 04264418e6f758dc777c33957fd652e048ef388bff51e5b84d505027fead1ca9
			AccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			Amount:         1000000000000,
			NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99,
				125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			NodeAddress:   "0.0.0.0",
			LockedBalance: 1000000,
		},
		{
			// 04266749faa93f9b6a15094c4d89037815455a76f254aeef2ebe4e445a538e0b
			AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			Amount:         1000000000000,
			NodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126,
				203, 5, 12, 152, 194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255},
			NodeAddress:   "0.0.0.0",
			LockedBalance: 1000000,
		},
		{
			// 04264a2ef814619d4a2b1fa3b45f4aa09b248d53ef07d8e92237f3cc8eb30d6d
			AccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			Amount:         1000000000000,
			NodePublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211, 123,
				72, 52, 221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
			NodeAddress:   "0.0.0.0",
			LockedBalance: 1000000,
		},
		{
			// Wallet Develop
			AccountAddress: "nK_ouxdDDwuJiogiDAi_zs1LqeN7f5ZsXbFtXGqGc0Pd",
			Amount:         10000000000,
			NodePublicKey: []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89,
				234, 181, 100, 163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220},
			NodeAddress:   "0.0.0.0",
			LockedBalance: 1000000,
		},
	}
	GenesisSignature = []byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
)
