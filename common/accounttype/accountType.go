package accounttype

// AccountType interface define the different behavior of each address
type (
	AccountType interface {
		// SetAccountPublicKey set/updated account public key
		SetAccountPublicKey(accountPublicKey []byte)
		// GetAccountAddress return the full (raw) account address in bytes
		GetAccountAddress() []byte
		// GetTypeInt return the value of the account address type in int
		GetTypeInt() uint32
		// GetAccountPublicKey return an account address in bytes
		GetAccountPublicKey() []byte
		// GetAccountPrefix return the value of current account address table prefix in the database
		GetAccountPrefix() string
		// GetName return the name of the account address type
		GetName() string
		// GetAccountPublicKeyLength return the length of this account address type (for parsing tx and message bytes that embed an address)
		GetAccountPublicKeyLength() uint32
		// IsEqual checks if two account have same type and pub key
		IsEqual(acc AccountType) bool
		// GetFormattedAccount return a string encoded/formatted account address
		// TODO: for this we have to move crypto signature to this package, otherwise we get circular dependency
		// GetFormattedAccount() (string, error)
	}
)
