package crypto

import (
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"golang.org/x/crypto/ed25519"
)

type (
	SignatureInterface interface {
		Sign(payload, accountID []byte, seed string) []byte
		SignBlock(payload []byte, nodeSeed string) []byte
		VerifySignature(payload, signature, accountID []byte) bool
	}

	// Signature object handle signing and verifying different signature
	Signature struct {
		Executor query.ExecutorInterface
	}
)

// NewSignature create new instance of signature object
func NewSignature(executor query.ExecutorInterface) *Signature {
	return &Signature{
		Executor: executor,
	}
}

// Sign accept account ID and payload to be signed then return the signature byte based on the
// signature method associated with account.Type
func (sig *Signature) Sign(payload, accountID []byte, seed string) []byte {
	accountQuery := query.NewAccountQuery()
	accountRows, _ := sig.Executor.ExecuteSelect(accountQuery.GetAccountByID(), accountID)
	var accounts []*model.Account
	account := accountQuery.BuildModel(accounts, accountRows)
	if len(account) == 0 {
		return nil
	}
	switch account[0].AccountType {
	case 0: // zoobc
		accountPrivateKey := ed25519GetPrivateKeyFromSeed(seed)
		signature := ed25519.Sign(accountPrivateKey, payload)
		return signature
	default:
		accountPrivateKey := ed25519GetPrivateKeyFromSeed(seed)
		signature := ed25519.Sign(accountPrivateKey, payload)
		return signature
	}
}

// SignBlock special method for signing block only, there will be no multiple signature options
func (*Signature) SignBlock(payload []byte, nodeSeed string) []byte {
	nodePrivateKey := ed25519GetPrivateKeyFromSeed(nodeSeed)
	return ed25519.Sign(nodePrivateKey, payload)
}

// VerifySignature accept payload (before without signature), signature and the account id
// then verify the signature + public key against the payload based on the
func (*Signature) VerifySignature(payload, signature, accountID []byte) bool {
	// todo: Fetch account from accountID
	account := &model.Account{
		ID: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139,
			255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		AccountType: 0,
		Address:     "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
	}

	switch account.AccountType {
	case 0: // zoobc
		result := ed25519.Verify(accountID, payload, signature)
		return result
	default:
		result := ed25519.Verify(accountID, payload, signature)
		return result
	}
}
