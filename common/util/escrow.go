package util

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

// ValidateBasicEscrow performs basic validation on the escrow
func ValidateBasicEscrow(tx *model.Transaction) error {
	if tx.Escrow == nil {
		return blocker.NewBlocker(blocker.ValidationErr, "InvalidEscrowObject")
	}
	if tx.Escrow.GetApproverAddress() == nil || bytes.Equal(tx.Escrow.GetApproverAddress(), []byte{}) {
		return blocker.NewBlocker(blocker.ValidationErr, "ApproverAddressRequired")
	}
	if tx.Escrow.GetTimeout() < tx.GetTimestamp() {
		return blocker.NewBlocker(blocker.ValidationErr, "TimeoutHasPassed")
	}
	return nil
}

// PrepareEscrowObjectForAction prepares the escrow object to be processed by actions
func PrepareEscrowObjectForAction(tx *model.Transaction) *model.Escrow {
	tx.Escrow.ID = tx.ID
	tx.Escrow.SenderAddress = tx.SenderAccountAddress
	tx.Escrow.BlockHeight = tx.Height
	tx.Escrow.Latest = true
	return tx.Escrow
}
