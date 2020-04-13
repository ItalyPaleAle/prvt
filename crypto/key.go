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

package crypto

import (
	"crypto/rand"
	"errors"

	"github.com/google/tink/go/subtle/kwp"
	"golang.org/x/crypto/argon2"
)

// KeyFromPassphrase returns the 32-byte key derived from a passphrase and a salt using Argon2id
// It also returns a "confirmation hash" that can be used to ensure the passphrase is correct
func KeyFromPassphrase(passphrase string, salt []byte) (key []byte, confirmationHash []byte, err error) {
	// Ensure the passphrase isn't empty and that the salt is 16-byte
	if passphrase == "" {
		return nil, nil, errors.New("empty passphrase")
	}
	if len(salt) != 16 {
		return nil, nil, errors.New("invalid salt")
	}

	// Derive the key using Argon2id
	// From the docs: "The draft RFC recommends[2] time=1, and memory=64*1024 is a sensible number. If using that amount of memory (64 MB) is not possible in some contexts then the time parameter can be increased to compensate."
	// Generate 64 bytes: the first 32 are the key, and the rest is the confirmation hash
	gen := argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 4, 64)
	return gen[0:32], gen[32:64], nil
}

// NewSalt generates a new, 16-byte salt, useful for Argon2id
func NewSalt() ([]byte, error) {
	return RandomBytes(16)
}

// NewKey generates a new, 32-byte key (unwrapped), suitable for AES-256
func NewKey() ([]byte, error) {
	return RandomBytes(32)
}

// RandomBytes returns a byte slice full with random bytes, of a given length
// This is useful to generate cryptographic keys, for example
func RandomBytes(len int) ([]byte, error) {
	key := make([]byte, len)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// WrapKey wraps a 32-byte key with another 32-byte key
func WrapKey(wrappingKey []byte, key []byte) ([]byte, error) {
	if len(wrappingKey) != 32 || len(key) != 32 {
		return nil, errors.New("keys must be 32-byte long")
	}

	// Get the key wrapper
	wrapper, err := kwp.NewKWP(wrappingKey)
	if err != nil {
		return nil, err
	}

	// Wrap the key
	return wrapper.Wrap(key)
}

// UnwrapKey unwraps a key wrapped with a 32-byte key
func UnwrapKey(wrappingKey []byte, wrappedKey []byte) ([]byte, error) {
	if len(wrappingKey) != 32 {
		return nil, errors.New("keys must be 32-byte long")
	}

	// Get the key wrapper
	wrapper, err := kwp.NewKWP(wrappingKey)
	if err != nil {
		return nil, err
	}

	// Unwrap the key
	return wrapper.Unwrap(wrappedKey)
}
