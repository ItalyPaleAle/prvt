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

// Metadata is the metadata for the fille
type Metadata struct {
	Name        string `json:"n,omitempty"`
	ContentType string `json:"ct,omitempty"`
	Size        int64  `json:"sz,omitempty"`
}

// MetadataCb is the callback for the function that receives the metadata object, as well as the length of the encoded metadata (including the size bytes)
type MetadataCb func(*Metadata, int)
