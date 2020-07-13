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

package utils

import (
	"errors"
	"net/textproto"
	"strconv"
	"strings"
)

/*
This file contains code adapted from:
https://github.com/golang/go/blob/0d20a492823211cd816ded24c98cfcd58b198faa/src/net/http/fs.go
Copyright 2009 The Go Authors. All rights reserved.
Licensed under a BSD-style license:
https://github.com/golang/go/blob/0d20a492823211cd816ded24c98cfcd58b198faa/LICENSE
*/

// HttpRange specifies the byte range to be sent to the client.
type HttpRange struct {
	Start, Length int64
}

// ParseRange parses a Range header string as per a subset of RFC 7233
// This supports two formats only: `bytes=startByte-` and `bytes=startByte-endByte`
// (that's the least common denominator supported by all services)
func ParseRange(s string) (*HttpRange, error) {
	// Check if the header is there
	if s == "" {
		return nil, nil
	}
	// Ensure the range is in bytes
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	// Ensure we have only one range
	if strings.Index(s[len(b):], ",") > -1 {
		return nil, errors.New("multiple ranges are not supported")
	}
	ra := textproto.TrimString(s[len(b):])
	if ra == "" {
		return nil, nil
	}
	i := strings.Index(ra, "-")
	if i < 0 {
		return nil, errors.New("invalid range")
	}
	start, end := textproto.TrimString(ra[:i]), textproto.TrimString(ra[i+1:])
	r := &HttpRange{}
	j, err := strconv.ParseInt(start, 10, 64)
	if err != nil || j < 0 {
		return nil, errors.New("invalid range")
	}
	r.Start = j
	if end == "" {
		// If no end is specified, range extends to end of the file
		r.Length = 0
	} else {
		j, err := strconv.ParseInt(end, 10, 64)
		if err != nil || r.Start > j {
			return nil, errors.New("invalid range")
		}
		r.Length = j - r.Start + 1
	}

	return r, nil
}
