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
	"os"
	"reflect"
	"strconv"

	"github.com/google/tink/go/kwp/subtle"
	"golang.org/x/crypto/argon2"
)

// Options for the Argon2 Key Derivation Function
type Argon2Options struct {
	// Options for argon2id
	Variant     string `json:"a2,omitempty"`
	Version     uint16 `json:"a2v,omitempty"`
	Memory      uint32 `json:"a2m,omitempty"`
	Iterations  uint32 `json:"a2t,omitempty"`
	Parallelism uint8  `json:"a2p,omitempty"`
}

// Setup sets the parameters for the key derivation function
func (o *Argon2Options) Setup() error {
	// Set variant and version
	o.Variant = "argon2id"
	o.Version = 0x13

	// Check if we have the tune function, which is included with a build tag only
	method := reflect.ValueOf(o).MethodByName("Tune")
	if method.IsValid() {
		method.Call([]reflect.Value{})
	}

	// Check if we have environmental variables to tune this
	iterations, err := strconv.ParseUint(os.Getenv("PRVT_ARGON2_ITERATIONS"), 10, 32)
	if err == nil && iterations > 0 {
		o.Iterations = uint32(iterations)
	}
	memory, err := strconv.ParseUint(os.Getenv("PRVT_ARGON2_MEMORY"), 10, 32)
	if err == nil && memory > 0 {
		o.Memory = uint32(memory)
	}
	parallelism, err := strconv.ParseUint(os.Getenv("PRVT_ARGON2_PARALLELISM"), 10, 8)
	if err == nil && parallelism > 0 {
		o.Parallelism = uint8(parallelism)
	}

	return o.Validate()
}

// Validate the values and sets the defaults for the missing ones
func (o *Argon2Options) Validate() error {
	// Default variant is argon2id, and only one supported
	if o.Variant == "" {
		o.Variant = "argon2id"
	} else if o.Variant != "argon2id" {
		return errors.New("unsupported variant for argon2")
	}
	// Default version is 19 (0x13), and only one supported
	if o.Version == 0 {
		o.Version = 0x13
	} else if o.Version != 0x13 {
		return errors.New("unsupported version for argon2")
	}
	// Ensure that this library uses version 0x13 too
	if argon2.Version != 0x13 {
		panic("argon2 library uses a different version that expected")
	}
	// Default memory is 80MB (in KB)
	if o.Memory == 0 {
		o.Memory = 80 * 1024
	}
	// Default iterations is 4
	if o.Iterations == 0 {
		o.Iterations = 4
	}
	// Default parallelism is 2
	if o.Parallelism == 0 {
		o.Parallelism = 2
	}
	return nil
}

// LegacyArgon2Options returns an object with the parameters for Argon2 used by prvt version 4 and below
func LegacyArgon2Options() *Argon2Options {
	return &Argon2Options{
		Variant:     "argon2id",
		Version:     19,
		Memory:      64 << 10,
		Iterations:  1,
		Parallelism: 4,
	}
}

// KeyFromPassphrase returns the 32-byte key derived from a passphrase and a salt using Argon2id
// It also returns a "confirmation hash" that can be used to ensure the passphrase is correct
func KeyFromPassphrase(passphrase string, salt []byte, kd *Argon2Options) (key []byte, confirmationHash []byte, err error) {
	// Ensure the passphrase isn't empty and that the salt is 16-byte
	if passphrase == "" {
		return nil, nil, errors.New("empty passphrase")
	}
	if len(salt) != 16 {
		return nil, nil, errors.New("invalid salt")
	}
	// Ensure the options are valid
	if kd == nil {
		kd = &Argon2Options{}
	}
	err = kd.Validate()
	if err != nil {
		return nil, nil, err
	}

	// Derive the key using Argon2id
	// Generate 64 bytes: the first 32 are the key, and the rest is the confirmation hash
	gen := argon2.IDKey([]byte(passphrase), salt, kd.Iterations, kd.Memory, kd.Parallelism, 64)
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
	wrapper, err := subtle.NewKWP(wrappingKey)
	if err != nil {
		return nil, err
	}

	// Wrap the key
	return wrapper.Wrap(key)
}

// UnwrapKey unwraps a key wrapped with a 32-byte key
func UnwrapKey(wrappingKey []byte, wrappedKey []byte) ([]byte, error) {
	if len(wrappingKey) != 32 {
		return nil, errors.New("wrapping key must be 32-byte long")
	}
	if len(wrappedKey) != 40 {
		return nil, errors.New("wrapped key must be 40-byte long")
	}

	// Get the key wrapper
	wrapper, err := subtle.NewKWP(wrappingKey)
	if err != nil {
		return nil, err
	}

	// Unwrap the key
	return wrapper.Unwrap(wrappedKey)
}
