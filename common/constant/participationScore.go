package constant

const (
	// ScalarReceiptScore the converter score to avoid calculation in float number, this value is following OneZBC to
	// maintain the number scale like balance does.
	ScalarReceiptScore = float32(OneZBC)
	// LinkedReceiptScore the score for each receipt that proved have relation with previous published receipt via merkle root
	LinkedReceiptScore float32 = 2
	// LinkedReceiptScore the score for each receipt that can't proved have relation with previous published receipt via merkle root
	UnlinkedReceiptScore float32 = 0.5
	// MaxScoreChange the maximum score that node wll get
	MaxScoreChange = 10 * int64(ScalarReceiptScore)
	// punishment amount
	ParticipationScorePunishAmount = -1 * MaxScoreChange / 2
	// MaxReceipt the maximum receipt will publish in every block
	MaxReceipt uint32 = 20
	// MaxParticipationScore maximum achievable score, this will be important to maintain smithing process so it doesn't
	// smith too fast
	MaxParticipationScore int64 = 10000000000 * int64(ScalarReceiptScore)
	// Starting score for newly registered nodes
	DefaultParticipationScore int64 = MaxParticipationScore / 10
)
