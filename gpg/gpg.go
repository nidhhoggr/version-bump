package gpg

import (
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

type EntityAccessorInterface interface {
	GetEntity(string, string) (*openpgp.Entity, error)
}

type EntityReaderInterface interface {
	ReadArmoredKeyRing(string) (openpgp.EntityList, error)
	GetPrivateKey(string, string) (string, error)
}

type EntityAccessor struct {
	Reader EntityReaderInterface
}

type EntityReader struct{}

func (ea *EntityAccessor) GetEntity(keyPassphrase string, signingKey string) (*openpgp.Entity, error) {
	privateKeyString, err := ea.Reader.GetPrivateKey(keyPassphrase, signingKey)
	if err != nil {
		return nil, err
	}
	es, err := ea.Reader.ReadArmoredKeyRing(privateKeyString)
	if err != nil {
		return nil, err
	}
	key := es[0]
	err = key.PrivateKey.Decrypt([]byte(keyPassphrase))
	return key, err
}

func (ea *EntityReader) ReadArmoredKeyRing(privateKey string) (openpgp.EntityList, error) {
	return openpgp.ReadArmoredKeyRing(strings.NewReader(privateKey))
}

func (ea *EntityReader) GetPrivateKey(keyPassphrase string, signingKey string) (string, error) {
	return getPrivateKey(keyPassphrase, signingKey)
}
