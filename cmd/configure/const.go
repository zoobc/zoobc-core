package configure

var (
	target string
	beta   = []string{
		"n0.beta.proofofparticipation.network:8002",
		"n1.beta.proofofparticipation.network:8002",
		"n2.beta.proofofparticipation.network:8002",
		"139.162.81.96:8002",
		"139.162.122.118:8002",
		"139.162.87.235:8002",
		"139.162.70.14:8002",
		"139.162.109.31:8002",
		"139.162.76.56:8002",
		"139.162.87.19:8002",
		"139.162.48.183:8002",
		"172.105.114.99:8002",
		"172.105.181.190:8002",
		"172.105.189.18:8002",
		"172.105.172.96:8002",
		"172.105.107.167:8002",
		"172.105.20.238:8002",
		"172.105.17.238:8002",
	}
	alpha = []string{
		"n0.alpha.proofofparticipation.network:8001",
		"n1.alpha.proofofparticipation.network:8001",
		"n2.alpha.proofofparticipation.network:8001",
		"172.105.37.61:8001",
		"80.85.84.163:8001",
	}
	dev = []string{
		"172.104.34.10:8001",
		"45.79.39.58:8001",
		"85.90.246.90:8001",
	}

	// maxAttemptPromptFailed the maximum allowed to try re-input prompt
	maxAttemptPromptFailed = 3
)
