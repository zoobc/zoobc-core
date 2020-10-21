package accounttype

import (
	"github.com/zoobc/zoobc-core/common/model"
)

// AccountTypeInterface interface define the different behavior of each address
type (
	AccountTypeInterface interface {
		// SetAccountPublicKey set/update account public key
		SetAccountPublicKey(accountPublicKey []byte)
		// GetAccountAddress return the full (raw) account address in bytes
		GetAccountAddress() ([]byte, error)
		// GetTypeInt return the value of the account address type in int
		GetTypeInt() int32
		// GetAccountPublicKey return an account address in bytes
		GetAccountPublicKey() []byte
		// GetAccountPrefix return the value of current account address table prefix in the database
		GetAccountPrefix() string
		// GetName return the name of the account address type
		GetName() string
		// GetAccountPublicKeyLength return the length of this account address type (for parsing tx and message bytes that embed an address)
		GetAccountPublicKeyLength() uint32
		// GetEncodedAddress return a string encoded/formatted account address
		GetEncodedAddress() (string, error)
		// GetAccountPublicKeyString return a string encoded account public key
		GetAccountPublicKeyString() (string, error)
		GetAccountPrivateKey() ([]byte, error)
		// IsEqual checks if two account have same type and pub key
		IsEqual(acc AccountTypeInterface) bool
		// GetSignatureType return the signature type number for this account type
		GetSignatureType() model.SignatureType
		// GetSignatureLength return the signature length for this account type
		GetSignatureLength() uint32

		// Sign accept a payload to be signed with an account seed then return the signature byte based on the
		Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error)
		VerifySignature(payload, signature, accountAddress []byte) error
		GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error
	}
)
