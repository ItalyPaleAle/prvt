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
	"strconv"
	"strings"
)

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
