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

// Header is the file header
type Header struct {
	Version uint16 `json:"v"`
	Key     []byte `json:"k"`
}

// Maximum length of the (encoded) metadata: 32KB - 2 bytes
// The first 2 bytes are used for the length
// This could be increased to up to 64KB - 2 bytes
const MaxMetadataLength = 32766

// MetadataCb is the callback for the function that receives the metadata object, as well as the length of the encoded metadata (including the size bytes)
type MetadataCb func(*Metadata, int32) bool
