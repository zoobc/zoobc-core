package accounttype

import (
	"github.com/zoobc/zoobc-core/common/model"
)

// AccountType interface define the different behavior of each address
type (
	AccountType interface {
		// SetAccountPublicKey set/update account public key
		SetAccountPublicKey(accountPublicKey []byte)
		// SetEncodedAccountAddress set/update encoded accountAddress (string representation of the accountAddress)
		// TODO: this should be calculated internally,
		//  but due to the difficulty in using crypto package inside this package for now we do it outside the scope of this interface
		SetEncodedAccountAddress(encodedAccount string)
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
		// GetFormattedAccount return a string encoded/formatted account address
		GetFormattedAccount() (string, error) // IsEqual checks if two account have same type and pub key
		IsEqual(acc AccountType) bool
		// GetSignatureType return the signature type number for this account type
		GetSignatureType() model.SignatureType
		// GetSignatureLength return the signature length for this account type
		GetSignatureLength() uint32

		// // GetAccountSignatureInterface return the signature implementation for this account type
		// // TODO: for now there is only one signature implementation that implements multiple signatures
		// GetAccountSignatureInterface() crypto.SignatureInterface
		// // GetSignatureTypeInterface return the signature type implementation for this account type
		// GetSignatureTypeInterface() crypto.SignatureTypeInterface

	}
)
