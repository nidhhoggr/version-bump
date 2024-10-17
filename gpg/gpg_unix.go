package gpg

import (
	"fmt"
	"os/exec"
)

func getPrivateKey(passphrase string, key string) (string, error) {
	return execPgpCommand(passphrase, key)
}

func execPgpCommand(passphrase string, key string) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("/usr/bin/gpg --armor --pinentry-mode=loopback --passphrase='%s' --export-secret-key=%s", passphrase, key))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
