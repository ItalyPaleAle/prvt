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

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomBytes(t *testing.T) {
	var (
		res1 []byte
		res2 []byte
		err  error
	)
	// Should return an amount of bytes that's different every time
	res1, err = RandomBytes(20)
	assert.NoError(t, err)
	assert.Len(t, res1, 20)
	res2, err = RandomBytes(20)
	assert.NoError(t, err)
	assert.Len(t, res2, 20)
	assert.True(t, !bytes.Equal(res1, res2))
}

func TestNewKey(t *testing.T) {
	// Should return 32 bytes
	res, err := NewKey()
	assert.NoError(t, err)
	assert.Len(t, res, 32)
}

func TestNewSalt(t *testing.T) {
	// Should return 16 bytes
	res, err := NewSalt()
	assert.NoError(t, err)
	assert.Len(t, res, 16)
}

func TestKeyFromPassphrase(t *testing.T) {
	var (
		key      []byte
		confHash []byte
		err      error
	)
	// Test from a table
	table := []struct {
		Salt       []byte
		Passphrase string
		Key        []byte
		ConfHash   []byte
	}{
		{
			Salt:       []byte("aaaaaaaaaaaaaaaa"),
			Passphrase: "hello world",
			Key:        decodeHex("0d8e68828d5242395fbee0cde9ab2e9c26907293c0a32a37f8282c8e6c06b3fc"),
			ConfHash:   decodeHex("e6493a58c2ca55d3911a4fa99a35339dae949dd69de265020f7443c91d3b3162"),
		},
		{
			Salt:       []byte("aaaaaaaaaaaaaaaa"),
			Passphrase: "m'illumino d'immenso",
			Key:        decodeHex("2efb520aa19ced975217c3b177d746575837822a3078422df49152c34882e261"),
			ConfHash:   decodeHex("edf8833939295d0dfc4a276cf2c147d7ba9bbd6e6e072304ef600ef6c7af1bb6"),
		},
		{
			Salt:       []byte("abababababababab"),
			Passphrase: "m'illumino d'immenso",
			Key:        decodeHex("9ce59682265c8f4e49cfb255ca8aad51f691a3399b0693ab64ddafddd41af34f"),
			ConfHash:   decodeHex("d68a5fc851d775c11a563530d0ef6846701a0bcda80d5e5b73463ea1f4ac01a7"),
		},
	}

	for _, el := range table {
		key, confHash, err = KeyFromPassphrase(el.Passphrase, el.Salt)
		assert.NoError(t, err)
		assert.Equal(t, el.Key, key)
		assert.Equal(t, el.ConfHash, confHash)
	}

	// Test errors
	key, confHash, err = KeyFromPassphrase("", []byte("abababababababab"))
	assert.EqualError(t, err, "empty passphrase")
	assert.Nil(t, key)
	assert.Nil(t, confHash)
	key, confHash, err = KeyFromPassphrase("foo", []byte("123"))
	assert.EqualError(t, err, "invalid salt")
	assert.Nil(t, key)
	assert.Nil(t, confHash)
}

func TestWrapKey(t *testing.T) {
	var (
		res []byte
		err error
	)

	// Test from a table
	table := []struct {
		WrappingKey []byte
		Key         []byte
		Expect      []byte
	}{
		// Test vectors from https://github.com/google/wycheproof/blob/master/testvectors/kwp_test.json
		{
			WrappingKey: decodeHex("38e1b1d075d9d852b9a6c01c8ff6965af01bac457a4e339ae3e1d7b2ffacc0cd"),
			Key:         decodeHex("80ad6820f1c90981e2ca42b817a345c1179d0a11d8e23a8adc0505e13d87295a"),
			Expect:      decodeHex("b63b7e0fec7e315816233db6758fd3e744b9f6a40862bdf866487e53bcb950d8b2649269e51b4475"),
		},
		{
			WrappingKey: decodeHex("c641f1689d81caa8ba37d895272240664054ed974cfffc40e6c5c0cad1b916c7"),
			Key:         decodeHex("3fd0ba19955e46749f54d88e99d080b7339d588fe612ec0f4021ca3ca2104270"),
			Expect:      decodeHex("837cfc316b49299edaf427e0988020ee876204b29d847669daab72c8660b0d860e9de3bd851198ff"),
		},
		{
			WrappingKey: decodeHex("aa0ab9d68ed4a04e723f81b44c0c88d0bcde7a80cfd476eb4b8836d9aa01ec4c"),
			Key:         decodeHex("57faa8766f6d6a0aa1cf643f857c150df5b31303b50af480e21c4b5e8c8a15d5"),
			Expect:      decodeHex("0e9e2e9aa34bbf973d67bc534ac86fc5b5a5f9da5f026866177894ec6077a5c84501510e1bf4afb3"),
		},
	}

	for _, el := range table {
		res, err = WrapKey(el.WrappingKey, el.Key)
		assert.NoError(t, err)
		assert.Equal(t, el.Expect, res)
	}

	// Test errors
	res, err = WrapKey(
		decodeHex("aabbcc"),
		decodeHex("57faa8766f6d6a0aa1cf643f857c150df5b31303b50af480e21c4b5e8c8a15d5"),
	)
	assert.EqualError(t, err, "keys must be 32-byte long")
	assert.Nil(t, res)
	res, err = WrapKey(
		decodeHex("aa0ab9d68ed4a04e723f81b44c0c88d0bcde7a80cfd476eb4b8836d9aa01ec4c"),
		decodeHex("aabbcc"),
	)
	assert.EqualError(t, err, "keys must be 32-byte long")
	assert.Nil(t, res)
}

func TestUnwrapKey(t *testing.T) {
	var (
		res []byte
		err error
	)

	// Test from a table
	table := []struct {
		WrappingKey []byte
		WrappedKey  []byte
		Expect      []byte
	}{
		// Test vectors from https://github.com/google/wycheproof/blob/master/testvectors/kwp_test.json
		{
			WrappingKey: decodeHex("38e1b1d075d9d852b9a6c01c8ff6965af01bac457a4e339ae3e1d7b2ffacc0cd"),
			WrappedKey:  decodeHex("b63b7e0fec7e315816233db6758fd3e744b9f6a40862bdf866487e53bcb950d8b2649269e51b4475"),
			Expect:      decodeHex("80ad6820f1c90981e2ca42b817a345c1179d0a11d8e23a8adc0505e13d87295a"),
		},
		{
			WrappingKey: decodeHex("c641f1689d81caa8ba37d895272240664054ed974cfffc40e6c5c0cad1b916c7"),
			WrappedKey:  decodeHex("837cfc316b49299edaf427e0988020ee876204b29d847669daab72c8660b0d860e9de3bd851198ff"),
			Expect:      decodeHex("3fd0ba19955e46749f54d88e99d080b7339d588fe612ec0f4021ca3ca2104270"),
		},
		{
			WrappingKey: decodeHex("aa0ab9d68ed4a04e723f81b44c0c88d0bcde7a80cfd476eb4b8836d9aa01ec4c"),
			WrappedKey:  decodeHex("0e9e2e9aa34bbf973d67bc534ac86fc5b5a5f9da5f026866177894ec6077a5c84501510e1bf4afb3"),
			Expect:      decodeHex("57faa8766f6d6a0aa1cf643f857c150df5b31303b50af480e21c4b5e8c8a15d5"),
		},
	}

	for _, el := range table {
		res, err = UnwrapKey(el.WrappingKey, el.WrappedKey)
		assert.NoError(t, err)
		assert.Equal(t, el.Expect, res)
	}

	// Test errors
	res, err = UnwrapKey(
		decodeHex("aabbcc"),
		decodeHex("0e9e2e9aa34bbf973d67bc534ac86fc5b5a5f9da5f026866177894ec6077a5c84501510e1bf4afb3"),
	)
	assert.EqualError(t, err, "wrapping key must be 32-byte long")
	assert.Nil(t, res)
	res, err = UnwrapKey(
		decodeHex("aa0ab9d68ed4a04e723f81b44c0c88d0bcde7a80cfd476eb4b8836d9aa01ec4c"),
		decodeHex("aabbcc"),
	)
	assert.EqualError(t, err, "wrapped key must be 40-byte long")
	assert.Nil(t, res)
}

func decodeHex(str string) []byte {
	res, _ := hex.DecodeString(str)
	return res
}
