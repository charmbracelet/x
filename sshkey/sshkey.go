package sshkey

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/crypto/ssh"
)

// Open reads the path, and parses the key.
func Open(keyPath string) (ssh.Signer, error) {
	pemBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("sshkey: %w", err)
	}
	return Parse(keyPath, pemBytes)
}

// Parse tries to parse the given PEM into a ssh.Signer.
// If the key is encrypted, it will ask for the passphrase.
// The 'identifier' is used to identify the key to the user when asking for the
// passphrase.
func Parse(identifier string, pemBytes []byte) (ssh.Signer, error) {
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if isPasswordError(err) {
		return askPassAndParse(identifier, pemBytes)
	}
	if err != nil {
		return nil, fmt.Errorf("sshkey: %w", err)
	}

	return signer, nil
}

func askPassAndParse(identifier string, pemBytes []byte) (ssh.Signer, error) {
	pass, err := ask(identifier)
	if err != nil {
		return nil, fmt.Errorf("sshkey: %w", err)
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(pass))
	if err != nil {
		return nil, fmt.Errorf("sshkey: %w", err)
	}
	return signer, nil
}

func isPasswordError(err error) bool {
	var kerr *ssh.PassphraseMissingError
	return errors.As(err, &kerr)
}

func ask(path string) (string, error) {
	var pass string
	if err := huh.Run(
		huh.NewInput().
			Inline(true).
			Value(&pass).
			Title(fmt.Sprintf("Enter the passphrase to unlock %q: ", path)).
			EchoMode(huh.EchoModePassword),
	); err != nil {
		return "", fmt.Errorf("sshkey: %w", err)
	}
	return pass, nil
}
