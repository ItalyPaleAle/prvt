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

/**
 * Contains utilities to work with request ranges
 * Port of the fsutils.RequestRange class from the Go code
 * 
 * @property {number} Start - Start of the range that is requested from the plaintext, in bytes 
 * @property {number} Length - Amount of data requested in plaintext, from the Start byte
 * @property {number} HeaderOffset - Size of the header added by prvt at the beginning of the file (encoded crypto.Header, including size bytes)
 * @property {number} MetadataOffset - Size of the encoded metadata object added at the beginning of the plaintext (encoded crypto.Metadata, including 2 size bytes)
 * @property {number} FileSize - File size, which acts as hard cap if set
 */
export class RequestRange {
    constructor(start, length) {
        this.Start = start
        this.Length = length
        this.HeaderOffset = 0
        this.MetadataOffset = 0
        this.FileSize = 0
    }

    /**
     * Sets the FileSize value and ensures that Start and Length don't overflow
     * @param {number} size - File size
     */
    SetFileSize(size) {
        this.FileSize = size
        if (this.FileSize < 1) {
            this.FileSize = 0
            return
        }
        if (this.Start > this.FileSize) {
            this.Start = this.FileSize
            this.Length = 0
        }
        else if (this.Length == 0 || this.Length > (this.FileSize-this.Start)) {
            this.Length = this.FileSize - this.Start
        }
    }

    /**
     * Returns the start package number
     * @returns {number} Start package number
     */
    StartPackage() {
        // This is rounded down always
        return Math.floor((this.Start + this.MetadataOffset) / (64 * 1024))
    }

    /**
     * Returns the end package number
     * @returns {number} End package number
     */
    EndPackage() {
        // Adding +1 to round up
        return Math.floor((this.Start + this.Length + this.MetadataOffset) / (64 * 1024)) + 1
    }

    /**
     * Returns the number of packages that need to be requested
     * @returns {number} Number of packages to request
     */
    LengthPackages() {
        return this.EndPackage() - this.StartPackage()
    }

    /**
     * Returns the start value in bytes
     * That's the start range for the request to the fs
     * @returns {number} Start offset in bytes
     */
    StartBytes() {
        return this.StartPackage() * (64 * 1024 + 32) + this.HeaderOffset
    }

    /**
     * Returns the end value in bytes
     * Thats the end range for the request to the fs
     * @returns {number} End offset in bytes
     */
    EndBytes() {
        return this.EndPackage() * (64 * 1024 + 32) + this.HeaderOffset
    }

    /**
     * Returns the number of bytes that need to be requested
     * @returns {number} Number of bytes to request
     */
    LengthBytes() {
        return this.LengthPackages() * (64 * 1024 + 32)
    }

    /**
     * Returns the number of bytes that need to be skipped from the beginning of the (decrypted) stream to match the requested range
     * @returns {number} Number of bytes to skip from the beginning of the (decrypted) stream
     */
    SkipBeginning() {
        return (this.Start + this.MetadataOffset) % (64 * 1024)
    }

    /**
     * Returns the value for the Range HTTP request reader, in bytes
     * @returns {string} Value of the request header
     */
    RequestHeaderValue() {
        return 'bytes=' + this.StartBytes() + '-' + (this.EndBytes() - 1)
    }

    /**
     * Returns the value for the Content-Range HTTP response reader, in bytes
     * @returns {string} Value of the header for responses
     */
    ResponseHeaderValue() {
        if (this.FileSize > 0) {
            return 'bytes ' + this.Start + '-' + (this.Start + this.Length - 1) + '/' + this.FileSize
        }
        else {
            return 'bytes ' + this.Start + '-' + (this.Start + this.Length - 1) + '/*'
        }
    }
}
