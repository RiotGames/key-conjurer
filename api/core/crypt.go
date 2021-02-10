package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
)

// PassThroughProvider is a CryptoProvider that performs no operations on its input
type PassThroughProvider struct{}

func (*PassThroughProvider) Encrypt(_ context.Context, input []byte) ([]byte, error) {
	return input, nil
}

func (*PassThroughProvider) Decrypt(_ context.Context, input []byte) ([]byte, error) {
	return input, nil
}

var _ CryptoProvider = &PassThroughProvider{}

// A CryptoProvider gives the user the ability to encrypt and decrypt bytes using secrets that are not aware to the caller.
type CryptoProvider interface {
	Encrypt(ctx context.Context, input []byte) ([]byte, error)
	Decrypt(ctx context.Context, input []byte) ([]byte, error)
}

// KeyConjurer's client passes credentials to the server, which then encrypts them using a server secret and hands them back to the client for storage and re-use.
// This is done because KeyConjurer's server implementation requires usage of the raw username and password to log a user in to either OneLogin or Okta.
// This was previously part of the cloudprovider module but the functionality has been removed from there because:
// - You might not necessarily want to use encryption if you are using AWS
// - You should not need to provide an encryption provider to use an alternative cloud service
// - Requiring all endpoints to reach out to KMS ahead of time made testing very difficult

// Crypto encrypts credentials using a given provider when handling them from a client connection
type Crypto struct {
	provider CryptoProvider
}

// NewCrypto creates a new Crypto with the given provider.
func NewCrypto(provider CryptoProvider) Crypto { return Crypto{provider} }

// Encrypt encrypts the given credentials and returns a string that is suitable to be stored on the client.
//
// Depending on the implementation, this may take a long time and it is recommended that a context with a deadline be provided.
func (c *Crypto) Encrypt(ctx context.Context, credentials Credentials) (string, error) {
	b, err := json.Marshal(credentials)
	if err != nil {
		return "", err
	}

	b, err = c.provider.Encrypt(ctx, b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// Decrypt decrypts the credentials stored within the given credentials object and updates it in place.
//
// If the credentials object is not encrypted, this is a no-op.
func (c *Crypto) Decrypt(ctx context.Context, credentials *Credentials) error {
	// TODO we probably want better sentinel errors so we can more easily return the appropriate status code
	if !credentials.Encrypted() {
		return nil
	}

	b, err := base64.StdEncoding.DecodeString(credentials.Password)
	if err != nil {
		return err
	}

	b, err = c.provider.Decrypt(ctx, b)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, credentials)
}
