// Package sshkey provides utilities for parsing SSH private keys with passphrase support.
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
	return doParse(identifier, pemBytes, ssh.ParsePrivateKey, ssh.ParsePrivateKeyWithPassphrase)
}

// ParseRaw tries to parse the given PEM into a private key.
// If the key is encrypted, it will ask for the passphrase.
// The 'identifier' is used to identify the key to the user when asking for the
// passphrase.
func ParseRaw(identifier string, pemBytes []byte) (any, error) {
	return doParse(identifier, pemBytes, ssh.ParseRawPrivateKey, ssh.ParseRawPrivateKeyWithPassphrase)
}

func doParse[T any](
	identifier string,
	pemBytes []byte,
	parse func(pemBytes []byte) (T, error),
	parseWithPass func(pemBytes, passphrase []byte) (T, error),
) (T, error) {
	result, err := parse(pemBytes)
	if isPassphraseMissing(err) {
		passphrase, err := ask(identifier)
		if err != nil {
			return result, fmt.Errorf("sshkey: %w", err)
		}
		result, err := parseWithPass(pemBytes, passphrase)
		if err != nil {
			return result, fmt.Errorf("sshkey: %w", err)
		}
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("sshkey: %w", err)
	}
	return result, nil
}

func isPassphraseMissing(err error) bool {
	var kerr *ssh.PassphraseMissingError
	return errors.As(err, &kerr)
}

func ask(path string) ([]byte, error) {
	var pass string
	if err := huh.Run(
		huh.NewInput().
			Inline(true).
			Value(&pass).
			Title(fmt.Sprintf("Enter the passphrase to unlock %q: ", path)).
			EchoMode(huh.EchoModePassword),
	); err != nil {
		return nil, fmt.Errorf("sshkey: %w", err)
	}
	return []byte(pass), nil
}
