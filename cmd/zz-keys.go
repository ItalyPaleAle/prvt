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
	"strconv"
	"strings"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/fs"

	"github.com/manifoldco/promptui"
)

var gpgPath string

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

// NewInfoFile generates a new info file with a brand-new master key, wrapped either with a passphrase-derived key, or with GPG
func NewInfoFile(gpgKey string) (info *fs.InfoFile, errMessage string, err error) {
	// First, create the info file
	info, err = fs.InfoCreate()
	if err != nil {
		return nil, "Error creating info file", err
	}

	// Generate the master key
	masterKey, err := crypto.NewKey()
	if err != nil {
		return nil, "Error generating the master key", err
	}

	// Add the key
	errMessage, err = AddKey(info, masterKey, gpgKey)
	if err != nil {
		info = nil
	}

	return info, "", nil
}

// UpgradeInfoFile upgrades an info file from version 1 to 2
func UpgradeInfoFile(info *fs.InfoFile) (errMessage string, err error) {
	// Can only upgrade info files version 1
	if info.Version != 1 {
		return "Unsupported repository version", errors.New("This repository has already been upgraded or is using an unsupported version")
	}

	// GPG keys are already migrated into the Keys slice
	// But passphrases need to be migrated
	if len(info.Salt) > 0 && len(info.ConfirmationHash) > 0 {
		// Prompt for the passphrase to get the current master key
		passphrase, err := PromptPassphrase()
		if err != nil {
			return "Error getting passphrase", err
		}

		// Get the current master key from the passphrase
		masterKey, confirmationHash, err := crypto.KeyFromPassphrase(passphrase, info.Salt)
		if err != nil || bytes.Compare(info.ConfirmationHash, confirmationHash) != 0 {
			return "Cannot unlock the repository", errors.New("Invalid passphrase")
		}

		// Create a new salt
		newSalt, err := crypto.NewSalt()
		if err != nil {
			return "Error generating a new salt", err
		}

		// Create a new wrapping key
		wrappingKey, newConfirmationHash, err := crypto.KeyFromPassphrase(passphrase, newSalt)
		if err != nil {
			return "Error deriving the wrapping key", err
		}

		// Wrap the key
		wrappedKey, err := crypto.WrapKey(wrappingKey, masterKey)
		if err != nil {
			return "Error wrapping the master key", err
		}

		// Add the key
		err = fs.InfoAddPassphrase(info, newSalt, newConfirmationHash, wrappedKey)
		if err != nil {
			return "Error adding the key", err
		}

		// Remove the old key
		info.Salt = nil
		info.ConfirmationHash = nil
	}

	// Update the version
	info.Version = 2

	return "", nil
}

// AddKey adds a key to an info file
// If the GPG Key is empty, will prompt for a passphrase
func AddKey(info *fs.InfoFile, masterKey []byte, gpgKey string) (errMessage string, err error) {
	var salt, confirmationHash, wrappedKey []byte

	// No GPG key specified, so we need to prompt for a passphrase first
	if gpgKey == "" {
		// Get the passphrase and derive the wrapping key, after generating a new salt
		passphrase, err := PromptPassphrase()
		if err != nil {
			return "Error getting passphrase", err
		}
		salt, err = crypto.NewSalt()
		if err != nil {
			return "Error generating a new salt", err
		}
		var wrappingKey []byte
		wrappingKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, salt)
		if err != nil {
			return "Error deriving the wrapping key", err
		}

		// Wrap the key
		wrappedKey, err = crypto.WrapKey(wrappingKey, masterKey)
		if err != nil {
			return "Error wrapping the master key", err
		}

		// Add the key
		err = fs.InfoAddPassphrase(info, salt, confirmationHash, wrappedKey)
		if err != nil {
			return "Error adding the key", err
		}
	} else {
		// Use GPG to wrap the master key
		wrappedKey, err = GPGEncrypt(masterKey, gpgKey)
		if err != nil {
			return "Error encrypting the master key with GPG", err
		}

		// Add the key
		err = fs.InfoAddGPGWrappedKey(info, gpgKey, wrappedKey)
		if err != nil {
			return "Error adding the key", err
		}
	}

	return "", nil
}

// isKeyIdPassphrase checks if a key ID is for a passphrase, and returns the index of the key
// Returns -1 otherwise
func isKeyIdPassphrase(keyId string) int {
	// Key IDs for passphrases start with "p:" and then have a number
	if !strings.HasPrefix(keyId, "p:") {
		return -1
	}

	passphraseId, err := strconv.Atoi(keyId[2:])
	if err != nil {
		return -1
	}

	return passphraseId
}

// RemoveKey removes a key from the info file
func RemoveKey(info *fs.InfoFile, keyId string) (errMessage string, err error) {
	found := false

	// Check if we're removing a passphrase
	passphraseId := isKeyIdPassphrase(keyId)
	if passphraseId >= 0 {
		// Iterate through the keys looking for the right one
		i := 0
		n := 0
		for _, k := range info.Keys {
			// Add all GPG keys
			if k.GPGKey != "" {
				info.Keys[n] = k
				n++
				continue
			}

			if i == passphraseId {
				found = true
			} else {
				info.Keys[n] = k
				n++
			}
			i++
		}

		// Truncate the slice
		info.Keys = info.Keys[:n]
	} else {
		// Iterate through the keys looking for the right one
		n := 0
		for _, k := range info.Keys {
			if k.GPGKey != "" && k.GPGKey == keyId {
				found = true
				continue
			}
			info.Keys[n] = k
			n++
		}

		// Truncate the slice
		info.Keys = info.Keys[:n]
	}

	if !found {
		return "Key not found", errors.New("Could not find a key with the given ID")
	}

	return "", nil
}

// GetMasterKey gets the master key, either deriving it from a passphrase, or from GPG
func GetMasterKey(info *fs.InfoFile) (masterKey []byte, keyId string, errMessage string, err error) {
	// Iterate through all the keys
	// First, try all keys that are wrapped with GPG
	for _, k := range info.Keys {
		if k.GPGKey == "" || len(k.MasterKey) == 0 {
			continue
		}
		// Try decrypting with GPG
		masterKey, err = GPGDecrypt(k.MasterKey)
		if err == nil {
			return masterKey, k.GPGKey, "", nil
		}
	}

	// No GPG key specified or unlocking with a GPG key was not successful
	// We'll try with passphrases; first, prompt for it
	passphrase, err := PromptPassphrase()
	if err != nil {
		return nil, "", "Error getting passphrase", err
	}

	// Check if we have a version 1 key, where the master key is directly derived from the passphrase
	if len(info.Salt) != 0 && len(info.ConfirmationHash) != 0 {
		var confirmationHash []byte
		masterKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, info.Salt)
		if err == nil && bytes.Compare(info.ConfirmationHash, confirmationHash) == 0 {
			return masterKey, "LegacyKey", "", nil
		}
	}

	// Try all version 2 keys that are wrapped with a key derived from the passphrase
	i := 0
	for _, k := range info.Keys {
		if k.GPGKey != "" || len(k.MasterKey) == 0 {
			continue
		}

		// Ensure we have the salt and confirmation hash
		if len(k.Salt) == 0 || len(k.ConfirmationHash) == 0 {
			continue
		}

		// Try this key
		var wrappingKey, confirmationHash []byte
		wrappingKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, k.Salt)
		if err == nil && bytes.Compare(k.ConfirmationHash, confirmationHash) == 0 {
			masterKey, err = crypto.UnwrapKey(wrappingKey, k.MasterKey)
			if err != nil {
				return nil, "", "Error while unwrapping the master key", err
			}
			return masterKey, "p:" + strconv.Itoa(i), "", nil
		}

		i++
	}

	// Tried all keys and nothing worked
	return nil, "", "Cannot unlock the repository", errors.New("Invalid passphrase")
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
	// Cached value
	if gpgPath != "" {
		return gpgPath, nil
	}

	// First, look for gpg2
	path, err := exec.LookPath("gpg2")
	if err == nil && len(path) > 0 {
		gpgPath = path
		return path, nil
	}

	// Try gpg
	path, err = exec.LookPath("gpg")
	if err == nil && len(path) > 0 {
		gpgPath = path
		return path, nil
	}

	// Couldn't find the binary
	return "", errors.New("could not find GPG binary in PATH")
}
