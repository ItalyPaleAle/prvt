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

package fsutils

import (
	"fmt"
	"testing"

	"github.com/ItalyPaleAle/prvt/utils"
	"github.com/stretchr/testify/assert"
)

func TestRanges(t *testing.T) {
	var rng *RequestRange
	var rngHeader *utils.HttpRange

	// Basic test
	rng = &RequestRange{
		Start:  0,
		Length: 1024,
	}
	rng.HeaderOffset = 72
	rng.MetadataOffset = 50
	rng.SetFileSize(600107)
	assert.Equal(t, uint32(0), rng.StartPackage())
	assert.Equal(t, uint32(1), rng.EndPackage())
	assert.Equal(t, uint32(1), rng.LengthPackages())
	assert.Equal(t, int64(72), rng.StartBytes())
	assert.Equal(t, int64(65640), rng.EndBytes())
	assert.Equal(t, int64(65568), rng.LengthBytes())
	assert.Equal(t, 50, rng.SkipBeginning())
	assert.Equal(t, "bytes=72-65639", rng.RequestHeaderValue())
	assert.Equal(t, "bytes 0-1023/600107", rng.ResponseHeaderValue())
	assert.Equal(t, "RequestRange{Start: 0, Length: 1024, FileSize: 600107, HeaderOffset: 72, MetadataOffset: 50}", fmt.Sprintf("%s", rng))

	// Remove file size
	rng.SetFileSize(0)
	assert.Equal(t, "bytes 0-1023/*", rng.ResponseHeaderValue())

	// From header
	// Still in one package
	rngHeader = &utils.HttpRange{
		Start:  60000,
		Length: 10,
	}
	rng = NewRequestRange(rngHeader)
	rng.HeaderOffset = 72
	rng.MetadataOffset = 50
	rng.SetFileSize(600107)
	assert.Equal(t, uint32(0), rng.StartPackage())
	assert.Equal(t, uint32(1), rng.EndPackage())
	assert.Equal(t, uint32(1), rng.LengthPackages())
	assert.Equal(t, int64(72), rng.StartBytes())
	assert.Equal(t, int64(65640), rng.EndBytes())
	assert.Equal(t, int64(65568), rng.LengthBytes())
	assert.Equal(t, 60050, rng.SkipBeginning())
	assert.Equal(t, "bytes=72-65639", rng.RequestHeaderValue())
	assert.Equal(t, "bytes 60000-60009/600107", rng.ResponseHeaderValue())

	// Start longer than file size
	rng.Start = 9000000
	rng.SetFileSize(600107)
	assert.Equal(t, 10333, rng.SkipBeginning())
	assert.Equal(t, "bytes=590184-655751", rng.RequestHeaderValue())
	assert.Equal(t, "bytes 600107-600106/600107", rng.ResponseHeaderValue())

	// Across packages 2
	rngHeader = &utils.HttpRange{
		Start:  65409,
		Length: 77,
	}
	rng = NewRequestRange(rngHeader)
	rng.HeaderOffset = 72
	rng.MetadataOffset = 50
	rng.SetFileSize(600107)
	assert.Equal(t, uint32(0), rng.StartPackage())
	assert.Equal(t, uint32(2), rng.EndPackage())
	assert.Equal(t, uint32(2), rng.LengthPackages())
	assert.Equal(t, int64(72), rng.StartBytes())
	assert.Equal(t, int64(131208), rng.EndBytes())
	assert.Equal(t, int64(131136), rng.LengthBytes())
	assert.Equal(t, 65459, rng.SkipBeginning())
	assert.Equal(t, "bytes=72-131207", rng.RequestHeaderValue())
	assert.Equal(t, "bytes 65409-65485/600107", rng.ResponseHeaderValue())

	// No ending, last package only
	rngHeader = &utils.HttpRange{
		Start:  600010,
		Length: 0,
	}
	rng = NewRequestRange(rngHeader)
	rng.HeaderOffset = 72
	rng.MetadataOffset = 50
	rng.SetFileSize(600107)
	assert.Equal(t, uint32(9), rng.StartPackage())
	assert.Equal(t, uint32(10), rng.EndPackage())
	assert.Equal(t, uint32(1), rng.LengthPackages())
	assert.Equal(t, int64(590184), rng.StartBytes())
	assert.Equal(t, int64(655752), rng.EndBytes())
	assert.Equal(t, int64(65568), rng.LengthBytes())
	assert.Equal(t, 10236, rng.SkipBeginning())
	assert.Equal(t, "bytes=590184-655751", rng.RequestHeaderValue())
	assert.Equal(t, "bytes 600010-600106/600107", rng.ResponseHeaderValue())

	// No ending, from first package
	rngHeader = &utils.HttpRange{
		Start:  60000,
		Length: 0,
	}
	rng = NewRequestRange(rngHeader)
	rng.HeaderOffset = 72
	rng.MetadataOffset = 50
	rng.SetFileSize(600107)
	assert.Equal(t, uint32(0), rng.StartPackage())
	assert.Equal(t, uint32(10), rng.EndPackage())
	assert.Equal(t, uint32(10), rng.LengthPackages())
	assert.Equal(t, int64(72), rng.StartBytes())
	assert.Equal(t, int64(655752), rng.EndBytes())
	assert.Equal(t, int64(655680), rng.LengthBytes())
	assert.Equal(t, 60050, rng.SkipBeginning())
	assert.Equal(t, "bytes=72-655751", rng.RequestHeaderValue())
	assert.Equal(t, "bytes 60000-600106/600107", rng.ResponseHeaderValue())

	// No ending, from non-first package
	rngHeader = &utils.HttpRange{
		Start:  70000,
		Length: 0,
	}
	rng = NewRequestRange(rngHeader)
	rng.HeaderOffset = 72
	rng.MetadataOffset = 50
	rng.SetFileSize(600107)
	assert.Equal(t, uint32(1), rng.StartPackage())
	assert.Equal(t, uint32(10), rng.EndPackage())
	assert.Equal(t, uint32(9), rng.LengthPackages())
	assert.Equal(t, int64(65640), rng.StartBytes())
	assert.Equal(t, int64(655752), rng.EndBytes())
	assert.Equal(t, int64(590112), rng.LengthBytes())
	assert.Equal(t, 4514, rng.SkipBeginning())
	assert.Equal(t, "bytes=65640-655751", rng.RequestHeaderValue())
	assert.Equal(t, "bytes 70000-600106/600107", rng.ResponseHeaderValue())
}
