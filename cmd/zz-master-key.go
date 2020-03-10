/*
Copyright © 2020 Alessandro Segala (@ItalyPaleAle)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs"

	"github.com/manifoldco/promptui"
)

// PromptPassphrase prompts the user for a passphrase
func PromptPassphrase() (string, error) {
	prompt := promptui.Prompt{
		Validate: func(input string) error {
			if len(input) < 1 {
				return errors.New("Passphrase must not be empty")
			}
			return nil
		},
		Label: "Passphrase",
		Mask:  '*',
	}

	key, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return key, err
}

// NewMasterKey generates a new master key, either deriving it from a passphrase, or encrypting it with GPG
func NewMasterKey(gpgKey string) (info *fs.InfoFile, errMessage string, err error) {
	var salt, confirmationHash, encryptedKey []byte

	// No GPG key specified, so we need to prompt for a passphrase
	if gpgKey == "" {
		// Get the passphrase and derive the master key, after generating a new salt
		passphrase, err := PromptPassphrase()
		if err != nil {
			return nil, "Error getting passphrase", err
		}
		salt, err = crypto.NewSalt()
		if err != nil {
			return nil, "Error generating a new salt", err
		}
		_, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, salt)
		if err != nil {
			return nil, "Error deriving the master key", err
		}
	} else {
		// Use GPG to encrypt the master key
		// First, generate the master key
		masterKey, err := crypto.NewKey()
		if err != nil {
			return nil, "Error generating the master key", err
		}

		// Encrypt with GPG
		encryptedKey, err = GPGEncrypt(masterKey, gpgKey)
		if err != nil {
			return nil, "Error encrypting the master key with GPG", err
		}
	}

	// Create the info file
	info, err = fs.InfoCreate(salt, confirmationHash, encryptedKey)
	if err != nil {
		return nil, "Error creating info file", err
	}

	return info, "", nil
}

// GetMasterKey gets the master key, either deriving it from a passphrase, or from GPG
func GetMasterKey(info *fs.InfoFile) (masterKey []byte, errMessage string, err error) {
	// No GPG key specified, so we need to prompt for a passphrase
	if len(info.EncryptedKey) == 0 {
		// Ensure we have the salt and confirmation hash
		if len(info.Salt) == 0 || len(info.ConfirmationHash) == 0 {
			return nil, "Salt and confirmation hash not present in the info file", errors.New("invalid info file")
		}
		// Get the passphrase and derive the master key
		passphrase, err := PromptPassphrase()
		if err != nil {
			return nil, "Error getting passphrase", err
		}
		var confirmationHash []byte
		masterKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, info.Salt)
		if bytes.Compare(info.ConfirmationHash, confirmationHash) != 0 {
			return nil, "Invalid passphrase", err
		}
	} else {
		// Decrypt the key with GPG
		masterKey, err = GPGDecrypt(info.EncryptedKey)
		if err != nil {
			return nil, "Error decrypting the master key with GPG", err
		}
	}

	return
}

// GPGEncrypt encrypts data using the GPG binary
func GPGEncrypt(in []byte, key string) (out []byte, err error) {
	return runGPG(in, "--output", "-", "--encrypt", "--recipient", key, "-")
}

// GPGDecrypt decrypts data using the GPG binary
func GPGDecrypt(in []byte) (out []byte, err error) {
	return runGPG(in, "--output", "-", "--decrypt", "-")
}

// runGPG runs the GPG command with the given flags
func runGPG(in []byte, flags ...string) (out []byte, err error) {
	// Get the GPG command
	path, err := lookupGPG()
	if err != nil {
		return
	}

	// Run GPG
	cmd := exec.Command(path, flags...)
	cmd.Stdin = bytes.NewReader(in)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	err = cmd.Run()
	if err != nil {
		return
	}
	return ioutil.ReadAll(&outBuf)
}

// lookupGPG returns the path of the GPG binary
func lookupGPG() (string, error) {
	// First, look for gpg2
	path, err := exec.LookPath("gpg2")
	if err == nil && len(path) > 0 {
		return path, nil
	}

	// Try gpg
	path, err = exec.LookPath("gpg")
	if err == nil && len(path) > 0 {
		return path, nil
	}

	// Couldn't find the binary
	return "", errors.New("could not find GPG binary in PATH")
}
