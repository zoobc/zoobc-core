// Constants to generate/validate genesis block
package constant

const (
	MainchainGenesisBlockID int64 = 8730856812106333607
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
			AccountAddress: "GOyhJUaHmc473WupBhmDsLE949ytd-yRL0k8-Miqufqb",
			AccountBalance: 500000000,
		},
		{
			AccountAddress: "NNfrBJvWIavOSzZfr-Izwg5HAB4783QGX77ymtiBw9Me",
			AccountBalance: 200000000,
		},
		{
			AccountAddress: "28nnMqpVdpvmdy4wM199G7StXZ3pFn7-8HBqIZqKF61b",
			AccountBalance: 4924985001,
		},
		{
			AccountAddress: "8ECRi0d4dgRH-QFQyB0UXol9gOmERdffcE4GFriSuDJp",
			AccountBalance: 74995000,
		},
		{
			AccountAddress: "Kv0AsDeQyvb1WH1N1-yu19En0h4Tj6xXMo_XN5RTjPaC",
			AccountBalance: 1000000000,
		},
		{
			AccountAddress: "8gdmIJaXNnmWGT5vNtYEA6SM9i3V0DTx78KCpok3RA3R",
			AccountBalance: 500000000,
		},
		{
			AccountAddress: "Ui76BsWscDvcyYkTvOdA8e-f8-YdflgIgaoPu_ZQhzsQ",
			AccountBalance: 700000000,
		},
		{
			AccountAddress: "JBh5yYlaaELf_kTdGsBrgGaSZqoaj9lp9MpsZwRkuU7B",
			AccountBalance: 1000060000,
		},
		{
			AccountAddress: "VhhDV2NrSOTCKf8OcKL8CyQedEwefDD2STSvagU3ajxN",
			AccountBalance: 300000800,
		},
		{
			AccountAddress: "Pf354gQDXIJYbs2-fNk6qvu_p_KRe8ZT3I2b6dmAYm8T",
			AccountBalance: 1100000000,
		},
		{
			AccountAddress: "FjZEDzBhFKdbXluipYl6lgMOcmBYXHnX-XTZEY4l80QH",
			AccountBalance: 1300000000,
		},
		{
			AccountAddress: "eADaxgIbFwRsVu8MdxFbOy_lG9Qpb2ui7yZrcjKQ8Pdu",
			AccountBalance: 1000000000,
		},
		{
			AccountAddress: "iY_Mtm8aI2SUYupbyJqEHqyl_duaFfNMoSHh_1uXNeav",
			AccountBalance: 1000000000,
		},
		{
			AccountAddress: "Cv8myOXkydKhQ6OMlBYvGLdUWXPMY-KpF4LNWgakUR_K",
			AccountBalance: 100000000000,
		},
		{
			AccountAddress: "bZSMHSo-JPyB81bTHOK5kSrhKFd-ruP0Jw6i93GwrsUG",
			AccountBalance: 200000000000,
		},
		{
			AccountAddress: "j_bsf8Dt0_Qchqhcn2RJV279w-3EMLwvTb-Bo3GY-1Qv",
			AccountBalance: 89993001,
		},
		{
			AccountAddress: "mz1KVJRc34dat8uwPsBG_Beplqhz1gvN379kL5yDtQXB",
			AccountBalance: 10200000,
		},
		{
			AccountAddress: "_BJN_2bipb0NmuzGq6uVadTZn1hz_gzf7ejxAsnA_d3c",
			AccountBalance: 300360001,
		},
		{
			AccountAddress: "yGmmP7UTukBeUbFGgWeMH6ilegw0GwEkgLx5Uh4nZEtT",
			AccountBalance: 978799985001,
		},
		{
			AccountAddress: "XmKTGJTRgGtrEHNbZihN7FdV4fuH4A12xj0OjOj88OiW",
			AccountBalance: 31100000000,
		},
		{
			AccountAddress: "vJm7NcRAlB9ePHvuGzVtnNDseB0C6nLJxfchkurX4xDz",
			AccountBalance: 1400100000,
		},
		{
			AccountAddress: "dJ5rHm51hSIDX5RtLAgOzzfBXV4JB2X1Zt9jvRX2aLA-",
			AccountBalance: 102000,
		},
		{
			AccountAddress: "RSQlbkdVqGh0BIzo-KwTS1YlEECEkw3cbT0JWwMnbhwj",
			AccountBalance: 900,
		},
		{
			AccountAddress: "c6utWMw13QQfB0wg0yn5wBAjR5DFWSA9XusFHRAlIm8C",
			AccountBalance: 100000,
		},
		{
			AccountAddress: "iKmae-mbpDHAyUNNAZkhvQ4d3ffDbIzMS8eBG-UCG9k_",
			AccountBalance: 3710115004,
		},
		{
			AccountAddress: "j9j4M0ZkfWnXf8spUDxq1ripdbmyc9etzDC5hQ0g4yvl",
			AccountBalance: 4219297105,
		},
		{
			AccountAddress: "KmbCem8BSws610Ba4-DBxWKT5v53zPW1if7DiUoaRlgh",
			AccountBalance: 16000,
		},
		{
			AccountAddress: "W_GQ5tIe7Bo7hCu8tot4fgYbP9qJN6DRXPZbFW4joRAE",
			AccountBalance: 24004,
		},
		{
			AccountAddress: "rfLsTvyuBf-IvSX9Nod3O1Nmph4NcIMWY416sC8kaWwy",
			AccountBalance: 5000000000,
		},
		{
			AccountAddress: "RIwnKZz2UwV3q5cf2gFY7llMT3jUFXTjcO7LWdbWcy19",
			AccountBalance: 25469146205,
		},
		{
			AccountAddress: "nK_ouxdDDwuJiogiDAi_zs1LqeN7f5ZsXbFtXGqGc0Pd",
			AccountBalance: 2920045458734,
			NodePublicKey: []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89, 234, 181, 100,
				163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220},
			NodeAddress:        "0.0.0.0",
			LockedBalance:      0,
			ParticipationScore: 100000000000000000,
		},
		{
			AccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			AccountBalance: 4183750658734,
			NodePublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211, 123, 72, 52,
				221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
			NodeAddress:        "0.0.0.0",
			LockedBalance:      0,
			ParticipationScore: 100000000000000000,
		},
		{
			AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			AccountBalance: 4183750658734,
			NodePublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152,
				194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255},
			NodeAddress:        "0.0.0.0",
			LockedBalance:      0,
			ParticipationScore: 100000000000000000,
		},
		{
			AccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			AccountBalance: 13638753743730,
			NodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118,
				97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			NodeAddress:        "0.0.0.0",
			LockedBalance:      10000000000000,
			ParticipationScore: 100000000000000000,
		},
	}
)
