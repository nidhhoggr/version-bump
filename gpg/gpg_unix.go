package gpg

import (
	"fmt"
	"os/exec"
)

func getPrivateKey(passphrase string, key string) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("gpg --armor --pinentry-mode=loopback --passphrase='%s' --export-secret-key=%s", passphrase, key))
	output, err := cmd.Output()
	return string(output), err
}
