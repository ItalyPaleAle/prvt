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

package infofile

import (
	"errors"
	"fmt"
)

const UnknownGPGKey = "UnknownGPGKey"

// InfoFileKey is a key that can be used to unlock a repo
type InfoFileKey struct {
	// Wrapped master key
	MasterKey []byte `json:"m,omitempty"`
	// Passphrase
	Salt             []byte `json:"s,omitempty"`
	ConfirmationHash []byte `json:"p,omitempty"`
	// GPG key id
	GPGKey string `json:"g,omitempty"`
}

// InfoFile is the content of the info file
type InfoFile struct {
	App      string `json:"app"`
	Version  uint16 `json:"ver"`
	DataPath string `json:"dp,omitempty"`

	// Fields for version 1
	// Passphrase
	Salt             []byte `json:"slt,omitempty"`
	ConfirmationHash []byte `json:"ph,omitempty"`
	// GPG-encrypted key
	EncryptedKey []byte `json:"ek,omitempty"`

	// Fields for version 2
	Keys []InfoFileKey `json:"k,omitempty"`
}

// New creates a new info file
func New() (*InfoFile, error) {
	info := &InfoFile{
		App:      "prvt",
		Version:  3,
		DataPath: "data",
	}
	return info, nil
}

// AddPassphrase adds a passphrase to an info file
func (info *InfoFile) AddPassphrase(salt []byte, confirmationHash []byte, wrappedKey []byte) error {
	key := InfoFileKey{
		Salt:             salt,
		ConfirmationHash: confirmationHash,
		MasterKey:        wrappedKey,
	}

	if info.Keys == nil {
		info.Keys = []InfoFileKey{}
	}

	info.Keys = append(info.Keys, key)
	return nil
}

// AddGPGWrappedKey adds a GPG-wrapped key to an info file
func (info *InfoFile) AddGPGWrappedKey(gpgKey string, wrappedKey []byte) error {
	key := InfoFileKey{
		GPGKey:    gpgKey,
		MasterKey: wrappedKey,
	}

	if info.Keys == nil {
		info.Keys = []InfoFileKey{}
	}

	info.Keys = append(info.Keys, key)
	return nil
}

// RemoveKey removes a key from the info file
func (info *InfoFile) RemoveKey(keyId string) error {
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
		return errors.New("Could not find a key with the given ID")
	}

	return nil
}

// Validate validates the info object
func (info *InfoFile) Validate() error {
	// Check the contents
	if info == nil {
		return errors.New("empty info object")
	}
	if info.App != "prvt" {
		return errors.New("invalid app name in info file")
	}

	if info.Version == 1 {
		// Parse version 1
		if info.EncryptedKey != nil {
			if len(info.EncryptedKey) < 30 {
				return errors.New("invalid encrypted master key")
			}

			// Convert the key to the slice as used by version 2
			key := InfoFileKey{
				MasterKey: info.EncryptedKey,
				GPGKey:    UnknownGPGKey,
			}
			info.Keys = []InfoFileKey{key}
			info.EncryptedKey = nil
		} else {
			// In version 1, the master key is directly derived from the passphrase
			if len(info.Salt) != 16 {
				return errors.New("invalid salt in info file")
			}
			if len(info.ConfirmationHash) != 32 {
				return errors.New("invalid confirmation hash in info file")
			}
		}
	} else if info.Version == 2 || info.Version == 3 {
		// Parse version 2 and 3
		if len(info.Keys) == 0 {
			return errors.New("repository does not have any key")
		}
		for i, k := range info.Keys {
			if len(k.MasterKey) < 30 {
				return fmt.Errorf("invalid wrapped master key for key %d", i)
			}
			if k.GPGKey == "" {
				if len(k.Salt) != 16 {
					return fmt.Errorf("invalid salt in info file for key %d", i)
				}
				if len(k.ConfirmationHash) != 32 {
					return fmt.Errorf("invalid confirmation hash in info file for key %d", i)
				}
			}
		}
	} else {
		// Unsupported version
		return errors.New("unsupported info file version – you might be trying to access a repository created with a newer version of prvt")
	}

	return nil
}
