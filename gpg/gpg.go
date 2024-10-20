package gpg

import (
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/input"
	"github.com/go-git/go-git/v5/config"
	"github.com/pkg/errors"
)

func GetSigningKeyFromConfig(gitConfig *config.Config) (string, error) {

	shouldNotSign, gpgVerificationKey := getSigningKeyFromConfig(gitConfig)

	if !shouldNotSign && gpgVerificationKey == "" {
		gitConfig, err := config.LoadConfig(config.GlobalScope)
		if err != nil {
			return "", errors.Wrap(err, "error loading git configuration from global scope")
		}
		_, gpgVerificationKey = getSigningKeyFromConfig(gitConfig)
	}

	return gpgVerificationKey, nil
}

func PromptForPassphrase(signingKey string) (*openpgp.Entity, error) {
	keyPassphrase, err := prompt.New().Ask("Input your passphrase:").
		Input("", input.WithEchoMode(input.EchoPassword))
	if err != nil {
		return nil, err
	}
	return getGpgEntity(keyPassphrase, signingKey)
}

func getSigningKeyFromConfig(config *config.Config) (bool, string) {

	var gpgVerificationKey string
	shouldNotSign := false

	commitSection := config.Raw.Section("commit")
	//logrus.Info(commitSection)
	if commitSection != nil && commitSection.Options.Get("gpgsign") == "true" {
		//logrus.Info(commitSection.Options.Get("gpgsign"))
		userSection := config.Raw.Section("user")
		//logrus.Info(userSection)
		if userSection != nil {
			//logrus.Info(userSection.Options.Get("signingkey"))
			gpgVerificationKey = userSection.Options.Get("signingkey")
		}
	} else if commitSection != nil && commitSection.Options.Get("gpgsign") == "false" {
		shouldNotSign = true
	}

	return shouldNotSign, gpgVerificationKey
}

func getGpgEntity(keyPassphrase string, signingKey string) (*openpgp.Entity, error) {

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
