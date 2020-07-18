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

// RequestRange is used to request a range of data only
// This function uses both bytes and packages, as generated by minio/sio; each package is 64KB + 32 bytes
type RequestRange struct {
	// Start of the range that is requested from the plaintext, in bytes
	Start int64
	// Amount of data requested in plaintext, from the Start byte
	Length int64
	// Size of the header added by prvt at the beginning of the file (encoded crypto.Header, including size bytes)
	HeaderOffset int64
	// Size of the encoded metadata object added at the beginning of the plaintext (encoded crypto.Metadata, including 2 size bytes)
	MetadataOffset int64
	// File size, which acts as hard cap if set
	FileSize int64
}

func (c *RequestRange) String() string {
	return fmt.Sprintf(
		"RequestRange{Start: %d, Length: %d, FileSize: %d, HeaderOffset: %d, MetadataOffset: %d}",
		c.Start,
		c.Length,
		c.FileSize,
		c.HeaderOffset,
		c.MetadataOffset,
	)
}

// NewRequestRange returns a new RequestRange object from a HttpRange one
func NewRequestRange(rng *utils.HttpRange) *RequestRange {
	return &RequestRange{
		Start:  rng.Start,
		Length: rng.Length,
	}
}

// SetFileSize sets the FileSize value and ensures that Start and Length don't overflow
func (c *RequestRange) SetFileSize(size int64) {
	c.FileSize = size
	if c.FileSize < 1 {
		c.FileSize = 0
		return
	}
	if c.Start > c.FileSize {
		c.Start = c.FileSize
		c.Length = 0
	} else if c.Length < c.Start || (c.Start+c.Length) > c.FileSize {
		if c.Length > (c.FileSize - c.Start) {
			c.Length = c.FileSize - c.Start
		}
	}
}

// StartPackage returns the start package number
func (c *RequestRange) StartPackage() uint32 {
	// This is rounded down always
	return uint32(c.Start+c.MetadataOffset) / (64 * 1024)
}

// EndPackage returns the end package number
func (c *RequestRange) EndPackage() uint32 {
	// Adding +1 to round up
	return uint32(c.Start+c.Length+c.MetadataOffset)/(64*1024) + 1
}

// LengthPackages returns the number of packages that need to be requested
func (c *RequestRange) LengthPackages() uint32 {
	return c.EndPackage() - c.StartPackage()
}

// StartBytes returns the start value in bytes
// That's the start range for the request to the fs
func (c *RequestRange) StartBytes() int64 {
	return (int64(c.StartPackage()) * (64*1024 + 32)) + c.HeaderOffset
}

// EndBytes returns the end value in bytes
// Thats the end range for the request to the fs
func (c *RequestRange) EndBytes() int64 {
	return (int64(c.EndPackage()) * (64*1024 + 32)) + c.HeaderOffset
}

// LengthBytes returns the number of bytes that need to be requested
func (c *RequestRange) LengthBytes() int64 {
	return int64(c.LengthPackages()) * (64*1024 + 32)
}

// SkipBeginning returns the number of bytes that need to be skipped from the beginning of the (decrypted) stream
// to match the requested range
func (c *RequestRange) SkipBeginning() int {
	return int((c.Start + c.MetadataOffset) % (64 * 1024))
}

// HeaderValue returns the value for the Range HTTP reader, in bytes
func (c *RequestRange) HeaderValue() string {
	return fmt.Sprintf("bytes=%d-%d", c.StartBytes(), c.EndBytes())
}
