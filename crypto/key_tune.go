// +build tune

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

// This file contains experimental code that calculates the "optimal" parameters for Argon2 based on the current system
// The Tune method uses as much available memory, with the max of 1GB, and as many cores as available for parallelization, with a max of 6. It then determines the ideal cost (number of iterations) so a key derivation takes between 0.3 and 0.9 seconds.
// This way of determining the tuned parameters is described in section 4 of the Argon2 RFC (https://tools.ietf.org/html/draft-irtf-cfrg-argon2-12#section-4)

// To compile this code, build the app with the `tune` tag, e.g. `go build -tags tune .`
// If this code is compiled, the Tune function is automatically run every time a new key is added (as part of the Setup() method)

import (
	"runtime"
	"time"

	"github.com/mackerelio/go-osstat/memory"
	"golang.org/x/crypto/argon2"
)

// Tune sets values for memory, iterations and parallelism based on the performance of this system
func (o *Argon2Options) Tune() {
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

func (o *Argon2Options) timeExecution() int64 {
	testBytes := []byte("aaaaaaaaaaaaaaaa")
	start := time.Now()
	argon2.IDKey(testBytes, testBytes, o.Iterations, o.Memory, o.Parallelism, 64)
	return time.Since(start).Milliseconds()
}
