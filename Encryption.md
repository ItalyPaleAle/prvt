# prvt encryption

prvt stores files using strong end-to-end encryption.

The files are encrypted on the local machine before being sent to the cloud or to the target directory. To view the files, one would need the encryption key (ie. the passphrase or GPG private key), as well as the encrypted files and their index.

When viewing files using the web-based interface, files are downloaded and then decrypted locally by prvt's local server, before being sent to the browser in cleartext.

## How files are encrypted

Files in a repository are encrypted using the [minio/sio](https://github.com/minio/sio) library. This library implements the "Data At Rest Encryption" (DARE) 2.0 format, and it encrypts files using either the AES-256-GCM or the ChaCha20-Poly1305.

Both are strong algorithms that provide authenticated encryption, guaranteeing the confidentiality as well as the integrity of the data. Both use a 256-bit (32-byte) key. The DARE format guarantees that stored files are tamper-resistant too.

sio, and by extension prvt, can decrypt files encrypted with either algorithm. When encrypting files, sio will use AES-256-GCM if the machine supports AES hardware acceleration (e.g. a CPU with AES-NI instructions), or fall back to ChaCha20-Poly1305 otherwise.

prvt uses a unique, randomly-generated 256-bit key for each file (using Go's crypto/rand). The same key is never re-used for more than one file, thus offering all the security benefits of the DARE format, including resistance to tampering.

## Encrypted files' headers

Each file stored in the repository (except the `_info.json` file) is encrypted with the DARE format. As an extension, however, files are added two headers, one in plaintext, and one part of the ciphertext.

The structure of each file is:

- The first 2 bytes are the size of the file header (encoded as little-endian)
- The file header follows (maximum 254 bytes)
- The rest of the file is encrypted with the DARE format. When decrypted, it contains:
    - The first 2 bytes of the encrypted data are the size of the metadata header (encoded as little-endian)
    - The metadata header follows (maximum 1,022 bytes)
    - The rest of the data is the original file

Visually:

```text
+------------------------------------------------+
|                                                |
| 2 bytes: size of the file header               |
| n bytes: file header (max 254 bytes)           |
|                                                |
|                                                |
| ENCRYPTED CONTENT WITH DARE                    |
| +--------------------------------------------+ |
| |                                            | |
| | 2 bytes: size of the metadata header       | |
| | n bytes: metadata header (max 1,022 bytes) | |
| |                                            | |
| | Original file follows                      | |
| |                                            | |
| +--------------------------------------------+ |
|                                                |
+------------------------------------------------+
```

### File header

The file header is a JSON fragment that contains 2 keys:

- The version (`v`) of the algorithm used to encrypt the file. Currently, this is always `1`.
- The wrapped encryption key (`k`) used to encrypt the file. Read more below on how the key is wrapped. In the JSON fragment, the key is base64-encoded.

For example:

```json
{"v":1,"k":"xfwZyE+zPscRlJU/BMsqkSjJwjW4S+qR5UD3Ss40X/KTr63548dzAQ=="}
```

Because this header tells prvt how to decrypt the file, and it contains the (wrapped) key used to encrypt it, it is stored in plain-text. The key is wrapped (i.e. encrypted), however, so having this file alone won't let anyone else decrypt the data.

The file header is at most 254 bytes in length.

### Metadata header

The metadata header is another JSON fragment, but this is stored encrypted, part of the ciphertext. It contains up to 3 keys (all optional):

- The name (`n`) of the file, as stored.
- The content type (`ct`) of the file, which is its MIME type.
- The size (`sz`) of the file in bytes.

For example:

```json
{"n":"IMG0342.jpeg","ct":"image/jpeg","sz":3432845}
```

Because the content of the metadata header can be sensitive, it's stored encrypted to protect your privacy.

While the metadata header is always present, it might not contain all (or any) keys, and prvt is still be able to decrypt the file.

The metadata header is at most 1,022 bytes in length.

## Master key

As mentioned above, each file is encrypted with a unique key that is randomly generated. The key is then wrapped (i.e. encrypted) with a master key.

The master key is a 256-bit symmetric key. prvt uses that to wrap each file's key using AES, as per [RFC 5649](https://tools.ietf.org/html/rfc5649). prvt relies on a module from the [google/tink](https://github.com/google/tink) library ([google/tink/go/subtle/kwp](https://godoc.org/github.com/google/tink/go/subtle/kwp)) to perform the key wrapping and unwrapping.

Each repository has one and only one master key, which is wrapped and stored in the `_info.json` file.

Starting from prvt version 0.3, a repository can be unlocked with more than one passphrases and/or GPG keys. Because of that, the `_info.json` file contains the master key as many times as the number of passphrases and GPG keys that can unlock the repository. Each one of them is wrapped with a wrapping key and using the same algorithm described above (AES as per RFC 5649, with the google/tink library).

- When using a passphrase, the wrapping key is derived from a user-supplied passphrase (using Argon2id)
- When using a GPG key, the wrapping key is encrypted using the GPG tool

### Wrapping key derived from a passphrase

The default method of operation of prvt uses a passphrase to derive the wrapping key (which is in turn used to unwrap the master key). The user sets the passphrase when first initializing the repository or with the `prvt repo key add` command, and then they're prompted for that when invoking any command in the prvt CLI (e.g. `prvt add`, `prvt serve`, `prvt rm`, etc).

In this mode of operation, the wrapping key is a 256-bit symmetric key that is derived from the user's passphrase using the [Argon2](https://en.wikipedia.org/wiki/Argon2) algorithm, in the Argon2id variant.

prvt uses the [golang.org/x/crypto/argon2](https://golang.org/x/crypto/argon2) implementation of Argon2id, which is part of the Go project. As per recommendation by the documentation (which itself references the [draft RFC](https://tools.ietf.org/html/draft-irtf-cfrg-argon2-03#section-9.3)), prvt uses Argon2id with time=1 and memory=64MB.

When deriving the wrapping key with Argon2id, prvt uses a random 16-byte salt, which is unique for each repository, and it's stored in cleartext in the info file (more on that below).

> When the wrapping key is derived from the passphrase, it's important to choose a passphrase with enough entropy. See [this site](https://www.useapassphrase.com/) for more information on passphrases.

### Wrapping key encrypted with GPG

This mode of operation is enabled with the `--gpg` (or `-g`) flag for the `prvt repo init` and `prvt repo key add` commands.

The CLI generates a random 256-bit key, which is the wrapping key. This key is then encrypted by invoking the GPG utility, and using the public key specified in the `--gpg` flag. The encrypted key is stored in the info file, and on every invocation of a prvt command that requires reading or writing data, the key is decrypted again using the GPG utility.

In order to use this option, clients need to have GPG version 2 or higher installed, as an external utility available in the system's `PATH`. They also need to have at least one key pair (public and secret) imported in their GPG keyring.

## File names and index

Each file that users add to the repository is given a name which is a random UUID, and it's placed in the same folder (some stores might divide files in sub-folders based on the first characters of the UUID, such as the local filesystem store).

This is done to protect your privacy, by hiding the original name of the file and its path.

To map files back to the original paths, prvt uses an encrypted index file. This is the `_index` file in the repository, and it's encrypted using the same pipeline as the data files, and as such it contains the same headers too.

Decrypted, the `_index` file is a structured document encoded with Protocol Buffer that contains a dictionary with two main keys:

- The version of the index file. The latest version, used from prvt version 0.4, is `2`.
- A list of elements present in the repository. This is an array of objects, each containing up to four keys:
    - The original path of the file within the repository's tree (for example, `/folder/sub/file.jpeg`).
    - The UUID of the encrypted file in the repository, stored as binary data.
    - The date when the file was added to the repository, stored as UNIX timestamp.
    - The mime type of the file, as determined from its extension.

The `_index` file uses Protocol Buffers for encoding, so it's a binary file and not human-readable. The proto file defining the data structure is saved at [`index/index.proto`](/index/index.proto).

Thanks to this index, prvt can show a tree of all directories and files, and knows what encrypted document to request for each file.

### Files in the repository

Please note that prvt creates one encrypted file in the data folder (by default, `data`) for each document that is added to the repository; files are currently not chunked into multiple pieces.

Because of that, it is possible for attackers who have access to your store to understand the number of files stored in a prvt repository and their approximate decrypted size. Additionally, if the file system where encrypted files are stored supports timestamps, attackers can also understand when files were added to the repository.

Currently, these limitations are by design. prvt does not aim to protect against disclosing the number of files in the repository, their approximate size, and the date they were stored is currently. This is considered an acceptable risk.

The data itself remains securely encrypted, and knowing the metadata above does not provide any advantage to attackers trying to decrypt the contents of your repository.

## The `_info.json` file

The `_info.json` file is the only file in the repository that is not encrypted.

This file is a JSON document containing four keys:

- The name of the app (`app`) that created it. This is always `prvt`.
- The version (`ver`) of the info file. The latest value, from prvt version 0.4, is `3`.
- The data path (`dp`), which is the name of the sub-folder where the encrypted data is stored. The default value is `data`. (This value can't be set using the prvt CLI, but it's defined here to enable backwards compatibility with repositories created by previous versions of prvt.)
- The list of passphrases and keys (`k`).

The last item is an array of dictionaries that contain the wrapped master key.

### For passphrases

For master keys that are wrapped with wrapping keys derived from a passphrase, the key object contains three keys:

- The wrapped master key (`m`).
- The salt (`s`) for deriving the master key using Argon2id, a 16-byte sequence encoded as base64.
- The passphrase's confirmation hash (`p`), a 32-byte sequence encoded as base64.

For example, for a repository that allows only one passphrase to unlock it (the document below has been pretty-printed for clarity for this example only):

```json
{
  "app": "prvt",
  "ver": 2,
  "dp": "data",
  "k": [
    {
      "m": "6AS6jAiD3TBx3/65aB2bs62w9rmP3hjnDC75LtuOFa+mF/FecuHUUQ==",
      "s": "VZMz8W64B4Zyc1Bu5ZS0Zw==",
      "p": "vpJFfif+EFLOfsX3Nsa9lrRC5xUSheq3yz7/1drlZRg="
    }
  ]
}
```

The last element, the passphrase's confirmation hash, is used to ensure that users are typing the correct passphrase.

The confirmation hash is generated in the same invocation of Argon2 that generates the master key. The Argon2 function returns 64 bytes: the first 32 are the wrapping key (used to wrap the master key), and the remaining 32 are used as the passphrase's confirmation hash, stored in cleartext in the info file.

When users run any command that requires reading or writing encrypted data in the repository (such as `prvt add`, `prvt serve`, etc), prvt invokes Argon2 to generate the 64-byte sequence from the passphrase, and compares the last 32 bytes with the value of `p` in the key dictionary. If they're different, it means that the user typed the wrong passphrase, and prvt will not use that key.

### For GPG keys

For master keys that are encrypted using GPG, instead, the key object in the `_info.json` file contains two elements:

- The wrapped master key (`m`).
- The address of the public GPG key (`g`).

For example, a repository that has only one key allowed to unlock it and that key is encrypted with GPG, looks like this (the document below has been pretty-printed for clarity for this example only, and the encrypted key was truncated):

```json
{
  "app": "prvt",
  "ver": 2,
  "dp": "data",
  "k": [
    {
      "m": "hQIMAwAAAAAAAAAAAQ...",
      "g": "0x30F411E2"
    }
  ]
}
```
