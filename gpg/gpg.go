package gpg

import (
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

type EntityAccessorInterface interface {
	GetEntity(keyPassphrase string, signingKey string) (*openpgp.Entity, error)
}

type EntityAccessor struct{}

func (ea *EntityAccessor) GetEntity(keyPassphrase string, signingKey string) (*openpgp.Entity, error) {
	privateKeyString, err := getPrivateKey(keyPassphrase, signingKey)
	if err != nil {
		return nil, err
	}
	es, err := openpgp.ReadArmoredKeyRing(strings.NewReader(privateKeyString))
	if err != nil {
		return nil, err
	}
	return es[0], nil
}
