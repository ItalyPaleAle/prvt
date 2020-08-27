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
	"crypto/sha256"
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
		masterKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, info.Salt)
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

		// Try this key
		var wrappingKey, confirmationHash []byte
		wrappingKey, confirmationHash, err = crypto.KeyFromPassphrase(passphrase, k.Salt)
		if err == nil && subtle.ConstantTimeCompare(k.ConfirmationHash, confirmationHash) == 1 {
			masterKey, err = crypto.UnwrapKey(wrappingKey, k.MasterKey)
			if err != nil {
				return nil, "", "Error while unwrapping the master key", err
			}
			hash := sha256.Sum256(k.MasterKey)
			return masterKey, fmt.Sprintf("p:%X", hash[0:8]), "", nil
		}
	}

	// Tried all keys and nothing worked
	return nil, "", "Cannot unlock the repository", errors.New("Invalid passphrase")
}
