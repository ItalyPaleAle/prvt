# prvt

prvt lets you store files on the cloud or on local directories, protected with strong end-to-end encryption, and then conveniently view them within a web browser.

Currently, prvt supports out-of-the-box storing files on a local folder or on [Azure Blob Storage](https://docs.microsoft.com/en-us/azure/storage/blobs/storage-blobs-overview).

prvt is free software, released under GNU General Public License version 3.0.

![The prvt web-based file viewer](./screenshot.png)

## Installation

## Pre-compiled binaries

The easiest way to install prvt is to download a pre-compiled binary, available for Windows, macOS, and Linux. Check out the [Releases](https://github.com/ItalyPaleAle/prvt/releases) section.

### Running on macOS

The pre-compiled binary is not signed with an Apple developer certificate, and recent versions of macOS will refuse to run it. You can fix this by running:

```sh
# Use the path where you downloaded prvt too
xattr -rc /path/to/prvt
```

## With go get

You can also fetch prvt with `go get`:

```sh
go get -u github.com/ItalyPaleAle/prvt
```

## Using prvt

## Initialize the repository

Before you can use prvt, you need to initialize a repository. This is done with the `prvt initrepo` command:

```sh
prvt initrepo --store <string>
```

You will be prompted to set a passphrase, which will be used to encrypt and decrypt all files.

The store flag tells prvt where to keep your files. It's a string that starts with the name of the store, followed by a provider-specific configuration. 

Supported stores at the moment are:

- For **Azure Blob Storage**, use `azure:` followed by the name of the container, for example `azure:myfiles`. The container must already exist. Additionally, set the following environmental variables to authenticate with Azure Storage: `AZURE_STORAGE_ACCOUNT` with the storage account name, and `AZURE_STORAGE_ACCESS_KEY` with the storage account key.
- For storing on a **local folder**: use `local:` and the path to the folder (absolute or relative to the current working directory). For example: `local:/path/to/folder` or `local:subfolder-in-cwd`.

For example, to store files locally in a folder called "repo" (in the current working directory):

```sh
prvt initrepo --store local:repo
```

To store on Azure Blob Storage in a storage account called "mystorageacct" and in the "myrepo" container:

```sh
export AZURE_STORAGE_ACCOUNT=mystorageacct
export AZURE_STORAGE_ACCESS_KEY=...
prvt initrepo --store azure:myrepo
```

### Add files

You can now add files to the repository, using the `prvt add` command:

```sh
# You'll be prompted for the repository's passphrase
prvt add <file> [<file> ...] --store <string> --destination <string>
```

You can add multiple files and folders, which will be added recursively.

The destination flag is required and it's the path in the repo where you want your files to be added; it must begin with a slash (`/`).

For example, to add the folder "photos" from your desktop:

```sh
prvt add ~/photos --store local:repo --destination /
```

### View files in the browser

prvt offers a browser-based interface to view your (encrypted) files, by running a local server. You can start the server with:

```sh
# You'll be prompted for the repository's passphrase
prvt serve --store <string>
```

By default, the server starts at http://127.0.0.1:3129 You can configure what port the server listens on with the `--port` flag. If you want to enable remote clients to access the server, use the `--address 0.0.0.0` flag.

Your browser will try to display supported files within itself, such as photos, supported videos, PDFs, etc. When trying to open other kinds of files, you'll be prompted to download them.

### Delete files from the repo

You can remove files from the repo with:

```sh
# You'll be prompted for the repository's passphrase
prvt rm <path> --store <string>
```

Where the path is the path of the file or folder within the repo. To remove a file, specify its exact path. To remove a folder recursively, specify the name of the folder, ending with `/*`.

For example, to remove a single file:

```sh
prvt rm /photos/IMG_0311.jpeg --store local:repo
```

To remove an entire folder:

```sh
prvt rm /photos/* --store local:repo
```

Note: once deleted, files cannot be recovered.

## FAQ

### How does prvt encrypt my files?

prvt encrypts your files using strong, industry-standard ciphers, such as AES-256-GCM and ChaCha20-Poly1305. The encryption key is derived from the passphrase you choose using Argon2id.

Check out the [Encryption](./Encryption.md) document for detailed information.

### Does prvt encrypt the names of files and folders?

Yes. prvt stores all encrypted files with a random UUID as name. The name of the file and its directory are only stored in the index file, which is encrypted itself.

### Has the prvt codebase been audited?

The prvt codebase has not been audited yet (and you won't see a "1.0" release until that happens).

However, all the cryptographic operations used by prvt leverage popular, strong ciphers and algorithms such as AES-256-GCM, ChaCha20-Poly1305, and Argon2id. prvt relies on production-ready libraries that implement those algorithms, such as [minio/sio](https://github.com/minio/sio), [google/tink](github.com/google/tink), and the Go's standard library.

Check out the [Encryption](./Encryption.md) document for detailed information.

### How many files can I store in a repo?

There's no limit on the number of files you can store in a repo.

However, the way the index is implemented relies on a single file, which might make opening or updating the files in a repository slow when you have many (thousands) of files. If you are planning to store a very large number of files, consider splitting them into multiple repositories.
