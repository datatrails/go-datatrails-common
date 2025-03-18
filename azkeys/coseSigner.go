package azkeys

/**
 * coseSigner implements the cose.Signer interface:
 *
 *  	https://github.com/veraison/go-cose
 */

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/veraison/go-cose"
)

// KeyVaultCoseSigner is the azure keyvault client for interacting with keyvault keys
// using a cose.Signer interface
type KeyVaultCoseSigner struct {
	// "Contexts should not be stored inside a struct type, but instead
	// passed to each function that needs it."
	// In this case this struct is transient and request scoped and the interface to the
	// methods does not include a context (and is outside of our control being defined
	// in the cose package).  We need the context for the logging ang tracing span.
	ctx                context.Context
	keyName            string
	alg                keyvault.JSONWebKeySignatureAlgorithm
	key                keyvault.KeyBundle
	locationIdentifier string
	*KeyVault
}

// NewKeyVaultCoseSigner creates a new keyvault configuration that signs with ES384
// using the latest version of the named key.
func NewKeyVaultCoseSigner(ctx context.Context, keyName string, keyVaultURL string, opts ...KeyVaultOption) (*KeyVaultCoseSigner, error) {
	hasher := sha256.New()
	hasher.Write([]byte("azure-keyvault"))
	locationHash := hasher.Sum(nil)
	locationHashString := hex.EncodeToString(locationHash)[:6]

	kv := KeyVaultCoseSigner{
		// We remove any cancelation function from the context so that it is
		// safe to stash it.  We must then not use this context for onward
		// requests without adding a further timeout.
		ctx:     context.WithoutCancel(ctx),
		keyName: keyName,
		alg:     keyvault.ES384, // hardwired for the moment, add caller setting when needed
		KeyVault: &KeyVault{
			url: keyVaultURL,
			log: logger.Sugar,
		},
		locationIdentifier: locationHashString,
	}

	// KeyVault options - not KeyVaultCoseSignerOptions
	for _, opt := range opts {
		opt(kv.KeyVault)
	}

	key, err := kv.GetLatestKey(ctx, kv.keyName)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest key for %s: %w", kv.keyName, err)
	}

	if key.Key.Kid == nil {
		return nil, fmt.Errorf("key ID is nil")
	}

	kv.key = key

	return &kv, nil
}

// KeyLocation returns an identifier for the place where the key is stored, used by the
// KeyIdentifier implementation.
func (kv *KeyVaultCoseSigner) KeyLocation() string {
	return kv.locationIdentifier
}

// Algorithm gets the cose algorithm for the key
func (kv *KeyVaultCoseSigner) Algorithm() cose.Algorithm {
	return cose.AlgorithmES384
}

// KeyIdentifier returns the essential information to identify the key, apart from any platform
// specific format (i.e. the Azure URL.) It takes the form: <location>:<key name>/<key version>.
// The location helps us identify where this key is stored. In this case, its azure key vault
func (kv *KeyVaultCoseSigner) KeyIdentifier() string {
	return fmt.Sprintf("%s:%s/%s", kv.KeyLocation(), GetKeyName(*kv.key.Key.Kid), GetKeyVersion(*kv.key.Key.Kid))
}

// base64BEtoBigInt takes a URL Encoded, base 64 encoded big endian byte array
// and returns a big.Int
func base64BEtoBigInt(in string) (*big.Int, error) {
	s, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return nil, fmt.Errorf("unable to base64 decode string %s: %w", in, err)
	}
	i := new(big.Int)
	// As the byte array is big endian no conversion is needed
	i.SetBytes(s)
	return i, nil
}

// publicKey returns the public key for the given keyvualt key bundle
//
// NOTE: Only valid for ECDSA
func publicKey(key keyvault.KeyBundle) (*ecdsa.PublicKey, error) {
	if key.Key.X == nil || key.Key.Y == nil {
		return nil, fmt.Errorf("public key is nil")
	}

	X, err := base64BEtoBigInt(*key.Key.X)
	if err != nil {
		return nil, fmt.Errorf("unable to convert X %s: %w", *key.Key.X, err)
	}

	Y, err := base64BEtoBigInt(*key.Key.Y)
	if err != nil {
		return nil, fmt.Errorf("unable to convert Y %s: %w", *key.Key.Y, err)
	}

	var curve elliptic.Curve

	switch key.Key.Crv {
	case keyvault.P256:
		curve = elliptic.P256()
	case keyvault.P384:
		curve = elliptic.P384()
	case keyvault.P521:
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("failed to find ecdsa curve for public key")
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     X,
		Y:     Y,
	}, nil

}

// PublicKey returns the public key for this instance of CoseSignerKeyVault
// given the kid.
//
// NOTE: Only valid for ECDSA
func (kv *KeyVaultCoseSigner) PublicKey(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {

	key, err := kv.GetKeyByKID(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest key for %s: %w", kv.keyName, err)
	}

	return publicKey(key)

}

// LatestPublicKey returns the latest public key for this instance of CoseSignerKeyVault
//
// NOTE: Only valid for ECDSA
func (kv *KeyVaultCoseSigner) LatestPublicKey() (*ecdsa.PublicKey, error) {
	return publicKey(kv.key)
}

// Sign signs a given content
func (kv *KeyVaultCoseSigner) Sign(rand io.Reader, content []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(kv.ctx, 30*time.Second)
	defer cancel()

	signature, err := kv.KeyVault.HashAndSign(
		ctx,
		content,
		*kv.key.Key.Kid,
		kv.alg,
	)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to sign: %w", err)
	}

	return signature, nil
}
