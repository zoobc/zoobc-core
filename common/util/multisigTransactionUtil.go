package util

import (
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"golang.org/x/crypto/sha3"
)

type (
	MultisigTransactionUtilInterface interface {
		CheckMultisigComplete(txBody *model.MultiSignatureTransactionBody, txHeight uint32) (bool, error)
	}
	MultisigTransactionUtil struct {
		PendingTransactionService service.PendingTransactionServiceInterface
		PendingSignatureService   service.PendingSignatureServiceInterface
		MultisigInfoService       service.MultisigInfoServiceInterface
		TransactionUtil           transaction.UtilInterface
	}
)

func NewMultisigTransactionUtil(
	pendingTransactionService service.PendingTransactionServiceInterface,
	pendingSignatureService service.PendingSignatureServiceInterface,
	multisigInfoService service.MultisigInfoServiceInterface,
	transactionUtil transaction.UtilInterface,
) *MultisigTransactionUtil {
	return &MultisigTransactionUtil{
		PendingTransactionService: pendingTransactionService,
		PendingSignatureService:   pendingSignatureService,
		MultisigInfoService:       multisigInfoService,
		TransactionUtil:           transactionUtil,
	}
}

func (mtu *MultisigTransactionUtil) CheckMultisigComplete(body *model.MultiSignatureTransactionBody, txHeight uint32) (bool, error) {
	if len(body.UnsignedTransactionBytes) > 0 {
		var (
			numberOfSubmittedSigs uint32
			multisigInfo          *model.MultiSignatureInfo
		)
		txHash := sha3.Sum256(body.UnsignedTransactionBytes)
		innerTx, err := mtu.TransactionUtil.ParseTransactionBytes(body.UnsignedTransactionBytes, false)
		if err != nil {
			return false, blocker.NewBlocker(
				blocker.ValidationErr,
				"FailToParseTransactionBytes",
			)
		}
		if body.MultiSignatureInfo != nil {
			multisigInfo = body.MultiSignatureInfo
		} else {
			multisigInfo, err = mtu.MultisigInfoService.GetMultisigInfoByAddress(innerTx.SenderAccountAddress)
			if err != nil {
				return false, err
			}
		}
		if multisigInfo == nil {
			return false, nil // no error but information is not complete yet
		}
		pendingSigs, err := mtu.PendingSignatureService.GetPendingSignatureByTransactionHash(txHash[:])
		if err != nil {
			return false, err
		}
		if body.SignatureInfo != nil {
			numberOfSubmittedSigs = uint32(len(body.SignatureInfo.Signatures))
		}
		if uint32(len(pendingSigs))+numberOfSubmittedSigs >= multisigInfo.MinimumSignatures {
			// completed tx
			return true, nil
		}
	} else if body.MultiSignatureInfo != nil {
		var (
			pendingTx             *model.PendingTransaction
			numberOfSubmittedSigs uint32
		)
		multisigAddress := body.MultiSignatureInfo.MultisigAddress
		if body.SignatureInfo != nil {
			numberOfSubmittedSigs = uint32(len(body.SignatureInfo.Signatures))
		}
		pendingSigs, err := mtu.PendingSignatureService.GetPendingSignatureByTransactionHash(txHash[:])
		if err != nil {
			return false, err
		}
		if uint32(len(pendingSigs))+numberOfSubmittedSigs < body.MultiSignatureInfo.MinimumSignatures {
			// not-completed tx
			return false, nil
		}
		if len(body.UnsignedTransactionBytes) > 0 {
			txHash := sha3.Sum256(body.UnsignedTransactionBytes)
			pendingTx = &model.PendingTransaction{
				TransactionHash:  txHash[:],
				TransactionBytes: body.UnsignedTransactionBytes,
				Status:           model.PendingTransactionStatus_PendingTransactionPending,
				BlockHeight:      txHeight,
			}
		} else {

		}

	} else if body.SignatureInfo != nil {
		txHash := body.SignatureInfo.TransactionHash

	}
	return nil, nil
}
