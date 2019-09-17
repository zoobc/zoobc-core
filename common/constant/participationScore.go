package constant

const (
	// ScalarReceiptScore the converter score to avoid calculation in float number
	ScalarReceiptScore float32 = 1000000
	// LinkedReceiptScore the score for each receipt that proved have relation with previous published receipt via merkle root
	LinkedReceiptScore float32 = 2
	// LinkedReceiptScore the score for each receipt that can't proved have relation with previous published receipt via merkle root
	UnlinkedReceiptScore float32 = 0.5
	// MaxScoreChange the maximum score that node wll get
	MaxScoreChange int64 = 10 * int64(ScalarReceiptScore)
	// MaxReceipt the maximum receipt will publish in every block
	MaxReceipt uint32 = 20
	// MaxParticipationScore maximum achievable score
	MaxParticipationScore int64 = 1000000 * int64(ScalarReceiptScore)
	// Starting score for newly registered nodes
	DefaultParticipationScore int64 = MaxParticipationScore / 10
)
