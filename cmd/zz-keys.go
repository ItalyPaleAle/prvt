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
	"crypto/subtle"
	"errors"
	"io"
	"os"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"
	"github.com/ItalyPaleAle/prvt/keys"

	"github.com/gofrs/uuid"
	"github.com/manifoldco/promptui"
)

// Input stream for usage with promptui; this can be overridden by tests
var PromptuiStdin io.ReadCloser = os.Stdin

// Output stream for usage with promptui; this can be overridden by tests
var PromptuiStdout io.WriteCloser = os.Stdout

// PromptPassphrase prompts the user for a passphrase
func PromptPassphrase() (string, error) {
	prompt := promptui.Prompt{
		Validate: func(input string) error {
			if len(input) < 1 {
				return errors.New("Passphrase must not be empty")
			}
			return nil
		},
		Label:  "Passphrase",
		Mask:   '*',
		Stdin:  PromptuiStdin,
		Stdout: PromptuiStdout,
	}

	key, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return key, err
}

// NewInfoFile generates a new info file with a brand-new master key, wrapped either with a passphrase-derived key, or with GPG
func NewInfoFile(gpgKey string) (info *infofile.InfoFile, errMessage string, err error) {
	// First, create the info file
	info, err = infofile.New()
	if err != nil {
		return nil, "Error creating info file", err
	}

	// Generate the master key
	masterKey, err := crypto.NewKey()
	if err != nil {
		return nil, "Error generating the master key", err
	}

	// Add the key
	_, errMessage, err = AddKey(info, masterKey, gpgKey)
	if err != nil {
		return nil, "Error adding the key", err
	}

	return info, "", nil
}

// UpgradeInfoFile upgrades an info file to the latest version
func UpgradeInfoFile(info *infofile.InfoFile) (errMessage string, err error) {
	// Can only upgrade info files versions 1-3
	if info.Version < 1 || info.Version > 3 {
		return "Unsupported repository version", errors.New("This repository has already been upgraded or is using an unsupported version")
	}

	// Upgrade 1 -> 2
	if info.Version < 2 {
		errMessage, err = upgradeInfoFileV1(info)
		if err != nil {
			return errMessage, err
		}
	}

	// Upgrade 2 -> 3
	// Nothing to do here, as the change is just in the index file
	// However, we still want to update the info file so older versions of prvt won't try to open a protobuf-encoded index file
	/*if info.Version < 3 {
	}*/

	// Upgrade 3 -> 4
	if info.Version < 4 {
		// Generate a new UUID
		repoId, err := uuid.NewV4()
		if err != nil {
			return "Could not generate a new UUID for the repo", err
		}
		info.RepoId = repoId.String()
	}

	// Update the version
	info.Version = 4

	return "", nil
}

// Upgrade an info file from version 1 to 2
func upgradeInfoFileV1(info *infofile.InfoFile) (errMessage string, err error) {
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
		if err != nil || subtle.ConstantTimeCompare(info.ConfirmationHash, confirmationHash) == 0 {
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
		err = info.AddPassphrase(newSalt, newConfirmationHash, wrappedKey)
		if err != nil {
			return "Error adding the key", err
		}

		// Remove the old key
		info.Salt = nil
		info.ConfirmationHash = nil
	}

	return "", nil
}

// AddKey adds a key to an info file
// If the GPG Key is empty, will prompt for a passphrase
func AddKey(info *infofile.InfoFile, masterKey []byte, gpgKey string) (keyId string, errMessage string, err error) {
	if gpgKey == "" {
		// Prompt for a passphrase first
		passphrase, err := PromptPassphrase()
		if err != nil {
			return "", "Error getting passphrase", err
		}

		// Add the passphrase
		return keys.AddKeyPassphrase(info, masterKey, passphrase)
	} else {
		// Add the GPG key
		return keys.AddKeyGPG(info, masterKey, gpgKey)
	}
}

// GetMasterKey gets the master key, either unwrapping it with a passphrase or with GPG
func GetMasterKey(info *infofile.InfoFile) (masterKey []byte, keyId string, errMessage string, err error) {
	// First, try unwrapping the key using GPG
	masterKey, keyId, errMessage, err = keys.GetMasterKeyWithGPG(info)
	if err == nil {
		return
	}

	// No GPG key specified or unlocking with a GPG key was not successful
	// We'll try with passphrases; first, prompt for it
	passphrase, err := PromptPassphrase()
	if err != nil {
		return nil, "", "Error getting passphrase", err
	}

	// Try unwrapping using the passphrase
	return keys.GetMasterKeyWithPassphrase(info, passphrase)
}
