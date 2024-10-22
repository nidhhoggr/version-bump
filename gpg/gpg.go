package gpg

import (
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

func GetGpgEntity(keyPassphrase string, signingKey string) (*openpgp.Entity, error) {

	privateKeyString, err := getPrivateKey(keyPassphrase, signingKey)
	if err != nil {
		return nil, err
	}

	s := strings.NewReader(privateKeyString)
	es, err := openpgp.ReadArmoredKeyRing(s)
	if err != nil {
		return nil, err
	}
	key := es[0]

	// double-checking but probably unnecessary
	err = key.PrivateKey.Decrypt([]byte(keyPassphrase))
	if err != nil {
		return nil, err
	}

	return key, nil
}
