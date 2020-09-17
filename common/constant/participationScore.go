package constant

const (
	// ScalarReceiptScore the converter score to avoid calculation in float number, this value is following OneZBC to
	// maintain the number scale like balance does.
	ScalarReceiptScore = float32(OneZBC)
	// LinkedReceiptScore the score for each receipt that proved have relation with previous published receipt via merkle root
	LinkedReceiptScore float32 = 2
	// LinkedReceiptScore the score for each receipt that can't proved have relation with previous published receipt via merkle root
	UnlinkedReceiptScore float32 = 0.5
	// MaxScoreChange the maximum score that node wll get.
	// note that in small networks if this value is too high it will lead to nodes being expelled from registry quickly
	// in production 100000000 * int64(ScalarReceiptScore). reduce to 10 * int64(ScalarReceiptScore) to test with less than 10 nodes
	MaxScoreChange = 1000 * int64(ScalarReceiptScore)
	// punishment amount
	ParticipationScorePunishAmount = -1 * MaxScoreChange / 2
	// MaxReceipt the maximum receipt will publish in every block
	MaxReceipt uint32 = 20
	// MaxParticipationScore maximum achievable score, this will be important to maintain smithing process so it doesn't
	// smith too fast
	MaxParticipationScore int64 = 10000000000 * int64(ScalarReceiptScore)
	// Starting score for newly registered nodes
	DefaultParticipationScore int64 = MaxParticipationScore / 10
	// Starting score for pre seed nodes (registered at genesis)
	GenesisParticipationScore int64 = MaxParticipationScore / 2
	// BetaBlockBonus beta only additional participation score
	BetaBlockBonus = MaxParticipationScore / 420
	// BetaBlockBonusSkipMultiplier beta only additional participation score decrease
	BetaBlockBonusSkipMultiplier = 5
)
