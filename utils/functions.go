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
	"path"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

var sanitizePathRegexp *regexp.Regexp

// SanitizePath removes certain problematic characters from path names
func SanitizePath(path string) string {
	// Compile the regular expression if necessary
	if sanitizePathRegexp == nil {
		sanitizePathRegexp = regexp.MustCompile("[#%&{}<>*\\$:!'\"+@\x60|=]")
	}

	// Unicode normalization
	path = norm.NFKC.String(path)

	// Replace all back slashes with a forward slash
	path = strings.ReplaceAll(path, "\\", "/")

	// Sanitize the string
	path = sanitizePathRegexp.ReplaceAllString(path, "")

	return path
}

var mimeTypeRegex = regexp.MustCompile("^(application|audio|font|image|model|text|video)\\/([a-z0-9-+*.]+)")

// SanitizeMimeType sanitizes a mime type
func SanitizeMimeType(mime string) string {
	// Lowercase the string and trim whitespaces to start
	mime = strings.TrimSpace(strings.ToLower(mime))

	// Ensure the format is correct
	// Base reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
	mime = mimeTypeRegex.FindString(mime)

	return mime
}

// IsIgnoredFile returns true if the file should be ignored and not added to the repository
// Ignored files include OS metadata files etc
func IsIgnoredFile(file string) bool {
	base := path.Base(file)

	// Invalid paths
	if base == "" ||
		base == "/" ||
		// Hidden files on *nix systems and on macOS
		// This includes all macOS metadata (.DS_Store, .AppleDouble, .LSOverride, .Trash, and files starting with ._)
		// It also includes .directory on Linux
		base[0] == '.' ||
		// Linux temp files
		base[len(base)-1] == '~' ||
		// Windows
		base == "Thumbs.db" ||
		base == "Thumbs.db:encrypted" ||
		base == "desktop.ini" ||
		base == "Desktop.ini" {
		return true
	}

	return false
}

// IsTruthy returns true if a string represent a "true" value, such as "true", "1", etc
func IsTruthy(str string) bool {
	str = strings.ToLower(str)
	return str == "1" ||
		str == "true" ||
		str == "t" ||
		str == "y" ||
		str == "yes"
}
