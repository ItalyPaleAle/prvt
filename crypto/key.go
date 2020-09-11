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
	"runtime"
	"time"

	"github.com/google/tink/go/kwp/subtle"
	"github.com/mackerelio/go-osstat/memory"
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

// Tune sets values for memory, iterations and parallelism based on the performance of this system
func (o *Argon2Options) Tune() {
	// Set variant and version
	o.Variant = "argon2id"
	o.Version = 0x13

	// Get the number of cores and set parallelism to number of cores, with min of 1 and max of 6
	cores := runtime.NumCPU()
	if cores < 1 {
		o.Parallelism = 1
	} else if cores > 6 {
		o.Parallelism = 6
	} else {
		o.Parallelism = uint8(cores)
	}

	// Set the memory usage to the free memory available (rounded to 16MB), with a minimum of 64MB and a maximum of 1GB
	var mem uint64
	stat, err := memory.Get()
	if err == nil {
		// Convert to KB
		mem = stat.Free / 1024
	}
	if mem < 80<<10 {
		mem = 80 << 10
	} else if mem > 1<<30 {
		mem = 1 << 30
	} else {
		// Round to 16MB
		mem = mem - (mem % (16 << 10))
	}
	o.Memory = uint32(mem)

	// Test iterations
	o.Iterations = 1

	// Run the test to adjust iterations
	for {
		time := o.timeExecution()

		// If we're doing one single iteration and this is too slow already, decrease memory by 200MB
		if o.Iterations == 1 && time > 900 {
			// Min we'll go to is 80 MB
			o.Memory -= 200 << 10
			if o.Memory < 80<<10 {
				o.Memory = 80 << 10
				break
			}
		} else if time < 300 {
			// Increase iterations if this is too fast
			o.Iterations *= 2
			if o.Iterations > 2<<6 {
				// Limit to 128 iterations
				break
			}
		} else {
			// If it's still to slow, decrease iterations and accept that
			if o.Iterations > 1 && time > 900 {
				o.Iterations /= 2
				break
			}
			// If we're here, we found optimal parameters
			break
		}
	}
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
	// Default memory is 64MB (in KB), for backwards compatibility
	if o.Memory == 0 {
		o.Memory = 64 * 1024
	}
	// Default iterations is 1, for backwards compatibility
	if o.Iterations == 0 {
		o.Iterations = 1
	}
	// Default parallelism is 4, for backwards compatibility
	if o.Parallelism == 0 {
		o.Parallelism = 4
	}
	return nil
}

func (o *Argon2Options) timeExecution() int64 {
	testBytes := []byte("aaaaaaaaaaaaaaaa")
	start := time.Now()
	argon2.IDKey(testBytes, testBytes, o.Iterations, o.Memory, o.Parallelism, 64)
	return time.Since(start).Milliseconds()
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
		return nil, errors.New("keys must be 32-byte long")
	}

	// Get the key wrapper
	wrapper, err := subtle.NewKWP(wrappingKey)
	if err != nil {
		return nil, err
	}

	// Unwrap the key
	return wrapper.Unwrap(wrappedKey)
}
