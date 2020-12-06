#!/bin/sh

# Install NPM dependencies
(cd ui && npm ci)

# Add GPG key
gpg --import tests/fixtures/gpg-ci-key.asc
echo "$GPGKEY_ID_FULL:6:" | gpg --import-ownertrust
gpg --list-secret-keys

# Fix permissions
sudo chown vscode:vscode /home/vscode/prvt-data

# Create the "prvt" container in the Azure Storage emulator (Azurite)
# Authentication is through the AZURE_STORAGE_CONNECTION_STRING environmental variable
az storage container create --name prvt

# Configure Minio client to access the minio container, then create the "prvt" bucket
mc alias set minio http://$S3_ENDPOINT $AWS_ACCESS_KEY_ID $AWS_SECRET_ACCESS_KEY
mc mb minio/prvt
