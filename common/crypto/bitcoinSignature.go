package crypto

import (
	"hash"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
	"golang.org/x/crypto/sha3"
)

type (
	// BitcoinSignature represent of bitcoin signature
	BitcoinSignature struct {
		NetworkParams *chaincfg.Params
		Curve         *btcec.KoblitzCurve
	}
)

// DefaultBitcoinNetworkParams to return defult Bitcoin network params
func DefaultBitcoinNetworkParams() *chaincfg.Params {
	// MainNetParams have params that will can use to setup sepific format for bitcoin payment address
	// more:  https://en.bitcoin.it/wiki/Address
	// https://en.bitcoin.it/wiki/List_of_address_prefixes
	return &chaincfg.MainNetParams
}

// DefaultBitcoinCurve to return used bitcoin curve
func DefaultBitcoinCurve() *btcec.KoblitzCurve {
	// Bitcoin use a specific Koblitz curve secp256k1
	return btcec.S256()
}

// DefaultBitcoinPublicKeyFormat return recommended public key format
func DefaultBitcoinPublicKeyFormat() model.BitcoinPublicKeyFormat {
	// https://bitcoin.org/en/glossary/compressed-public-key
	return model.BitcoinPublicKeyFormat_PublicKeyFormatCompressed
}

// DefaultBitcoinPrivateKeyLength to
func DefaultBitcoinPrivateKeyLength() model.PrivateKeyBytesLength {
	return model.PrivateKeyBytesLength_PrivateKey256Bits
}

// NewBitcoinSignature is new instance of bitcoin signature
func NewBitcoinSignature(netParams *chaincfg.Params, curve *btcec.KoblitzCurve) *BitcoinSignature {
	return &BitcoinSignature{
		NetworkParams: netParams,
		Curve:         curve,
	}
}

// Sign to generates an ECDSA signature for the provided payload
func (*BitcoinSignature) Sign(privateKey *btcec.PrivateKey, payload []byte) ([]byte, error) {
	var sig, err = privateKey.Sign(payload)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.AuthErr, err.Error())
	}
	return sig.Serialize(), nil
}

// Verify to verify the signature of payload using provided public key
func (*BitcoinSignature) Verify(
	payload []byte,
	signature *btcec.Signature,
	publicKey *btcec.PublicKey,
) bool {
	return signature.Verify(payload, publicKey)
}

// GetNetworkParams to bitcoin network parameters
func (b *BitcoinSignature) GetNetworkParams() *chaincfg.Params {
	return b.NetworkParams
}

// GetPrivateKeyFromSeed to get private key form seed
func (b *BitcoinSignature) GetPrivateKeyFromSeed(
	seed string,
	privkeyLength model.PrivateKeyBytesLength,
) (*btcec.PrivateKey, error) {
	var (
		// Convert seed (secret phrase) to byte array
		seedBuffer = []byte(seed)
		hasher     hash.Hash
		privateKey *btcec.PrivateKey
	)
	switch privkeyLength {
	case model.PrivateKeyBytesLength_PrivateKey256Bits:
		hasher = sha3.New256()
	case model.PrivateKeyBytesLength_PrivateKey384Bits:
		hasher = sha3.New384()
	case model.PrivateKeyBytesLength_PrivateKey512Bits:
		hasher = sha3.New512()
	default:
		return nil, blocker.NewBlocker(blocker.AppErr, "invalidPrivateKeyLength")
	}
	if _, err := hasher.Write(seedBuffer); err != nil {
		return nil, err
	}
	privateKey, _ = btcec.PrivKeyFromBytes(b.Curve, hasher.Sum(nil))
	return privateKey, nil
}

// GetPublicKeyFromSeed Get the raw public key corresponding to a seed (secret phrase)
func (b *BitcoinSignature) GetPublicKeyFromSeed(
	seed string,
	format model.BitcoinPublicKeyFormat,
	privkeyLength model.PrivateKeyBytesLength,
) ([]byte, error) {
	var privateKey, err = b.GetPrivateKeyFromSeed(seed, privkeyLength)
	if err != nil {
		return nil, err
	}
	publicKey, err := b.GetPublicKeyFromPrivateKey(privateKey, format)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

// GetPublicKeyFromPrivateKey get raw public key from private key
// public key format : https://bitcoin.org/en/wallets-guide#public-key-formats
func (*BitcoinSignature) GetPublicKeyFromPrivateKey(
	privateKey *btcec.PrivateKey,
	format model.BitcoinPublicKeyFormat,
) ([]byte, error) {
	switch format {
	case model.BitcoinPublicKeyFormat_PublicKeyFormatUncompressed:
		return privateKey.PubKey().SerializeUncompressed(), nil
	case model.BitcoinPublicKeyFormat_PublicKeyFormatCompressed:
		return privateKey.PubKey().SerializeCompressed(), nil
	default:
		return nil, blocker.NewBlocker(blocker.AppErr, "invalidPublicKeyFormat")
	}
}

// GetPublicKeyFromBytes to get public key from raw bytes public key
func (b *BitcoinSignature) GetPublicKeyFromBytes(pubkey []byte) (*btcec.PublicKey, error) {
	return btcec.ParsePubKey(pubkey, b.Curve)
}

// GetPublicKeyString will return hex string from bytes public key
func (b *BitcoinSignature) GetPublicKeyString(publicKey []byte) (string, error) {
	var address, err = btcutil.NewAddressPubKey(publicKey, b.GetNetworkParams())
	if err != nil {
		return "", blocker.NewBlocker(blocker.ParserErr, err.Error())
	}
	return address.String(), nil
}

// GetAddressFromPublicKey to get address public key from seed
func (b *BitcoinSignature) GetAddressFromPublicKey(publicKey []byte) (string, error) {
	var address, err = btcutil.NewAddressPubKey(publicKey, b.GetNetworkParams())
	if err != nil {
		return "", blocker.NewBlocker(blocker.ParserErr, err.Error())
	}
	return address.EncodeAddress(), nil
}

// GetAddressBytes Get raw bytes of a string encoded address
func (b *BitcoinSignature) GetAddressBytes(encodedAddress string) ([]byte, error) {
	var decodedAddress, err = btcutil.DecodeAddress(encodedAddress, b.GetNetworkParams())
	if err != nil {
		return nil, blocker.NewBlocker(blocker.ParserErr, err.Error())
	}
	return decodedAddress.ScriptAddress(), nil
}

// GetSignatureFromBytes to get signature type from signature raw bytes
func (b *BitcoinSignature) GetSignatureFromBytes(signatureBytes []byte) (*btcec.Signature, error) {
	var signature, err = btcec.ParseSignature(signatureBytes, b.Curve)
	if err != nil {
		return nil, blocker.NewBlocker(blocker.ParserErr, err.Error())
	}
	return signature, nil
}
