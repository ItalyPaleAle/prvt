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

package keys

import (
	"crypto/subtle"
	"errors"
	"fmt"

	"github.com/ItalyPaleAle/prvt/crypto"
	"github.com/ItalyPaleAle/prvt/infofile"
)

// GetMasterKeyWithGPG gets the master key unwrapping it with GPG
func GetMasterKeyWithGPG(info *infofile.InfoFile) (masterKey []byte, keyId string, errMessage string, err error) {
	// Iterate through all the keys looking for those wrapped with GPG
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

	// No valid key found
	return nil, "", "Cannot unlock the repository", errors.New("Invalid GPG key")
}

// GetMasterKeyWithPassphrase gets the master key unwrapping it with a passphrase
// (Supports v1 keys too, which were directly derived from the passphrase)
func GetMasterKeyWithPassphrase(info *infofile.InfoFile, passphrase string) (masterKey []byte, keyId string, errMessage string, err error) {
	// Check if we have a version 1 key, where the master key is directly derived from the passphrase
	if len(info.Salt) != 0 && len(info.ConfirmationHash) != 0 {
		var confirmationHash []byte
		// Default KDF options
		kdfOptions := crypto.LegacyArgon2Options()
		masterKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, info.Salt, kdfOptions)
		if err == nil && subtle.ConstantTimeCompare(info.ConfirmationHash, confirmationHash) == 1 {
			return masterKey, "LegacyKey", "", nil
		}
	}

	// Try all version 2 keys that are wrapped with a key derived from the passphrase
	for _, k := range info.Keys {
		// Skip GPG keys
		if k.GPGKey != "" || len(k.MasterKey) == 0 {
			continue
		}

		// Ensure we have the salt and confirmation hash
		if len(k.Salt) == 0 || len(k.ConfirmationHash) == 0 {
			continue
		}

		// Ensure the key derivation function is "argon2"
		// For backwards compatibility, accept empty values too
		if k.KDF != "argon2" && k.KDF != "" {
			continue
		}
		// For backwards compatibility, create the KDF options if empty and set the values to the legacy ones
		if k.KDFOptions == nil {
			k.KDFOptions = crypto.LegacyArgon2Options()
		}
		// Skip invalid keys
		if k.KDFOptions.Validate() != nil {
			continue
		}

		// Try this key
		var wrappingKey, confirmationHash []byte
		wrappingKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, k.Salt, k.KDFOptions)
		if err == nil && subtle.ConstantTimeCompare(k.ConfirmationHash, confirmationHash) == 1 {
			masterKey, err = crypto.UnwrapKey(wrappingKey, k.MasterKey)
			if err != nil {
				return nil, "", "Error while unwrapping the master key", err
			}
			return masterKey, fmt.Sprintf("p:%X", k.MasterKey[0:8]), "", nil
		}
	}

	// Tried all keys and nothing worked
	return nil, "", "Cannot unlock the repository", errors.New("Invalid passphrase")
}

// AddKeyPassphrase adds a new wrapping with a passphrase
func AddKeyPassphrase(info *infofile.InfoFile, masterKey []byte, passphrase string) (keyId string, errMessage string, err error) {
	var salt, confirmationHash, wrappedKey []byte

	// Before adding the key, check if it's already there
	_, _, _, testErr := GetMasterKeyWithPassphrase(info, passphrase)
	if testErr == nil {
		return "", "Key already added", errors.New("This passphrase has already been added to the repository")
	}

	// Set up parameters for Argon2
	kdfOptions := &crypto.Argon2Options{}
	err = kdfOptions.Setup()
	if err != nil {
		return "", "Error setting up Argon2 parameters", err
	}

	// Derive the wrapping key, after generating a new salt
	salt, err = crypto.NewSalt()
	if err != nil {
		return "", "Error generating a new salt", err
	}
	var wrappingKey []byte
	wrappingKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, salt, kdfOptions)
	if err != nil {
		return "", "Error deriving the wrapping key", err
	}

	// Wrap the key
	wrappedKey, err = crypto.WrapKey(wrappingKey, masterKey)
	if err != nil {
		return "", "Error wrapping the master key", err
	}

	// Add the key
	err = info.AddPassphrase(salt, confirmationHash, wrappedKey, kdfOptions)
	if err != nil {
		return "", "Error adding the key", err
	}

	// Return the key ID
	return fmt.Sprintf("p:%X", wrappedKey[0:8]), "", nil
}

// AddKeyGPG adds a new wrapping with GPG
func AddKeyGPG(info *infofile.InfoFile, masterKey []byte, gpgKey string) (keyId string, errMessage string, err error) {
	var wrappedKey []byte

	// Normalize the key ID
	gpgKey = NormalizeGPGKeyId(gpgKey)
	if gpgKey == "" {
		return "", "Invalid GPG key", errors.New("GPG key ID is not in the correct format")
	}
	// Before adding the key, check if it's already there
	for _, k := range info.Keys {
		if k.GPGKey == gpgKey {
			return "", "Key already added", errors.New("This GPG key has already been added to the repository")
		}
	}

	// Use GPG to wrap the master key
	wrappedKey, err = GPGEncrypt(masterKey, gpgKey)
	if err != nil {
		return "", "Error encrypting the master key with GPG", err
	}

	// Add the key
	err = info.AddGPGWrappedKey(gpgKey, wrappedKey)
	if err != nil {
		return "", "Error adding the key", err
	}

	// Return the key ID
	return gpgKey, "", nil
}
