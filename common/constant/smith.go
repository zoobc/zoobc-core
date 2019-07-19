package constant

var (
	Two64                 = "18446744073709551616"
	MaximumBalance        = int64(10000000000)
	InitialSmithScale     = int64(153722867)
	MaxSmithScale         = InitialSmithScale * MaximumBalance
	GenesisBlocksmithID   = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	GenesisBlockSignature = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	GenesisBlockTimestamp = int64(1562117271)
	GenesisAccountAddress = "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7" // 042643fd5c5f3b64bd6b38882f7727fdc5d6eee3c354f3ab9d7b8ff7c1840140
)
