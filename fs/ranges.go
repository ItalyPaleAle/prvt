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

package fs

import (
	"fmt"

	"github.com/ItalyPaleAle/prvt/utils"
)

// PackageRange is used to request a range of packages only
// Each package is 64KB + 32 bytes
// The offset is for the file's header added by prvt (crypto.Header encoded)
type PackageRange struct {
	Start, Length, Offset int64
}

// NewPackageRange returns a new PackageRange object from a HttpRange one (which uses bytes)
func NewPackageRange(rngBytes utils.HttpRange) (c *PackageRange) {
	c = &PackageRange{}
	c.Start = rngBytes.Start / (64*1024 + 32)
	// Length is + 2 because of rounding up and because the first package is always smaller as it contains the metadata; it's best to fetch a bit more data (64KB) than not having all the data we need
	c.Length = (rngBytes.Length / (64*1024 + 32)) + 2
	return
}

// StartBytes returns the start value in bytes
func (c *PackageRange) StartBytes() int64 {
	return (c.Start * (64*1024 + 32)) + c.Offset
}

// EndBytes returns the end value in bytes
func (c *PackageRange) EndBytes() int64 {
	return ((c.Start + c.Length) * (64*1024 + 32)) + c.Offset - 1
}

// LengthBytes returns the length value in bytes
func (c *PackageRange) LengthBytes() int64 {
	return c.Length * (64*1024 + 32)
}

// HeaderValue returns the value for the Range HTTP reader, in bytes
func (c *PackageRange) HeaderValue() string {
	return fmt.Sprintf("bytes=%d-%d", c.StartBytes(), c.EndBytes())
}
