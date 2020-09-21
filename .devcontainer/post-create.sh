#!/bin/sh

# Install NPM dependencies
(cd ui && npm ci)

# Add GPG key
gpg --import tests/fixtures/gpg-ci-key.asc
echo "$GPGKEY_ID_FULL:6:" | gpg --import-ownertrust
gpg --list-secret-keys
