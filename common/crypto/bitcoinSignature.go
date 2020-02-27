package crypto

import (
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
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

// DefaultCurve to return used bitcoin curve
func DefaultCurve() *btcec.KoblitzCurve {
	// Bitcoin use a specific Koblitz curve secp256k1
	return btcec.S256()
}

// DefaultPublicKeyFormat return recomended public key format
func DefaultPublicKeyFormat() btcutil.PubKeyFormat {
	// https://bitcoin.org/en/glossary/compressed-public-key
	return btcutil.PKFCompressed
}

// NewBitcoinSignature is new instance of bitcoin signature
func NewBitcoinSignature(netParams *chaincfg.Params, curve *btcec.KoblitzCurve) *BitcoinSignature {
	return &BitcoinSignature{
		NetworkParams: netParams,
		Curve:         curve,
	}
}

// Sign to generates an ECDSA signature for the provided payload
func (b *BitcoinSignature) Sign(privateKey *btcec.PrivateKey, payload []byte) ([]byte, error) {
	var sig, err = privateKey.Sign(payload)
	if err != nil {
		return nil, err
	}
	return sig.Serialize(), nil
}

// Verify to verify the signature of payload using provided public key
func (b *BitcoinSignature) Verify(
	payload []byte,
	signature *btcec.Signature,
	publicKey *btcec.PublicKey,
) bool {

	return signature.Verify(payload, publicKey)
}

// GetNetworkParams to bitcoin network paramters
func (b *BitcoinSignature) GetNetworkParams() *chaincfg.Params {
	return b.NetworkParams
}

// GetPrivateKeyFromSeed to get private key form seed
func (b *BitcoinSignature) GetPrivateKeyFromSeed(seed string) *btcec.PrivateKey {
	var (
		// Convert seed (secret phrase) to byte array
		seedBuffer = []byte(seed)
		// Compute SHA3-256 hash of seed (secret phrase)
		seedHash      = sha3.Sum256(seedBuffer)
		privateKey, _ = btcec.PrivKeyFromBytes(b.Curve, seedHash[:])
	)
	return privateKey
}

// GetPublicKeyFromSeed Get the raw public key corresponding to a seed (secret phrase)
// public key format : https://bitcoin.org/en/wallets-guide#public-key-formats
func (b *BitcoinSignature) GetPublicKeyFromSeed(seed string, format btcutil.PubKeyFormat) []byte {
	var privateKey = b.GetPrivateKeyFromSeed(seed)
	switch format {
	case btcutil.PKFUncompressed:
		return privateKey.PubKey().SerializeUncompressed()
	case btcutil.PKFCompressed:
		return privateKey.PubKey().SerializeCompressed()
	case btcutil.PKFHybrid:
		return privateKey.PubKey().SerializeHybrid()
	default:
		return nil
	}
}

// GetAddressPublicKey to get address public key from seed
func (b *BitcoinSignature) GetAddressPublicKey(
	publicKey []byte,
) (string, error) {
	if publicKey != nil {
		return "", errors.New("Invalid Public Key")
	}
	var address, err = btcutil.NewAddressPubKey(publicKey, b.GetNetworkParams())
	if err != nil {
		return "", err
	}
	return address.String(), nil
}

// GetBytesAddressPublicKey Get row bytes of address
func (b *BitcoinSignature) GetBytesAddressPublicKey(address string) ([]byte, error) {
	var decodedAddress, err = btcutil.DecodeAddress(address, b.GetNetworkParams())
	if err != nil {
		return nil, err
	}
	return decodedAddress.ScriptAddress(), nil
}

// GetPublicKeyFromAddress to get public key from address
func (b *BitcoinSignature) GetPublicKeyFromAddress(address string) (*btcec.PublicKey, error) {
	rowBytesAddress, err := b.GetBytesAddressPublicKey(address)
	if err != nil {
		return nil, err
	}
	publicKey, err := btcec.ParsePubKey(rowBytesAddress, b.Curve)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

// GetSignatureFromBytes to get signature type from signature row bytes
func (b *BitcoinSignature) GetSignatureFromBytes(signatureBytes []byte) (*btcec.Signature, error) {
	var signature, err = btcec.ParseSignature(signatureBytes, b.Curve)
	if err != nil {
		return nil, err
	}
	return signature, nil
}
