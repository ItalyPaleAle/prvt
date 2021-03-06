APP_VERSION ?= dev
BUILD_TIME := $(shell date -u +'%Y-%m-%dT%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD)

.PHONY: all build test clean

# Performs a builld
all: build

# Fetches the tools that are required to build prvt
get-tools:
	mkdir -p .bin
	test -f .bin/go-acc || \
	  curl -sf https://gobinaries.com/github.com/ory/go-acc@v0.2.6 | PREFIX=.bin/ sh
	test -f .bin/pkger || \
	  curl -sf https://gobinaries.com/github.com/markbates/pkger/cmd/pkger@v0.17.1 | PREFIX=.bin/ sh

# Clean all compiled files
clean:
	rm -rfv ui/dist/* || true
	rm -rfv .bin/prvt* bin || true
	rm -v pkged.go || true

# Build the entire app
build: build-ui build-wasm-prod pkger build-app

# Build the Go code
build-app:
	test -f pkged.go || echo "WARN: pkger has not run"
	go build \
	  -ldflags "-X github.com/ItalyPaleAle/prvt/buildinfo.Production=1 -X github.com/ItalyPaleAle/prvt/buildinfo.AppVersion=$(APP_VERSION) -X github.com/ItalyPaleAle/prvt/buildinfo.BuildID=$(APP_VERSION) -X github.com/ItalyPaleAle/prvt/buildinfo.BuildTime=$(BUILD_TIME) -X github.com/ItalyPaleAle/prvt/buildinfo.CommitHash=$(COMMIT_HASH)" \
	  -o bin

# Build the web UI
build-ui:
	(cd ui; npm ci; APP_VERSION="$(APP_VERSION)" npm run build)

# Run pkger
pkger:
	.bin/pkger list
	.bin/pkger

# Copy the wasm_exec.js file from the Go installation
copy-wasm-runtime:
	# Copy the Go wasm runtime
	cp -v $$(go env GOROOT)/misc/wasm/wasm_exec.js ui/src/sw/

# Build the wasm binary
build-wasm:
	# Empty the directory
	rm ui/assets/*.wasm || true
	rm ui/assets/*.wasm.br || true
	# Build the wasm file
	( cd wasm; GOOS=js GOARCH=wasm go build -ldflags "-s -w" -o ../ui/assets/app.wasm )

# Build the wasm binary for production (compressed
build-wasm-prod: build-wasm
	# Compress with brotli
	brotli -j9 ui/assets/app.wasm

# Run tests
test: get-tools
	# Exclude the wasm package because it requires a different compilation target
	GPGKEY_ID="0x4C6D7DB1D92F58EE" \
	GPGKEY_USER="prvt CI <ci@prvt>" \
	  .bin/go-acc $(shell go list ./... | grep -v prvt/wasm) -- -v -ldflags "-X github.com/ItalyPaleAle/prvt/buildinfo.Production=1"
	# Remove generated (.pb.go ones) files from coverage report
	cat coverage.txt| grep -v ".pb.go:" > coverage-filtered.txt

# Run the shorter test suite
test-short: get-tools
	# Exclude the wasm package because it requires a different compilation target
	GPGKEY_ID="0x4C6D7DB1D92F58EE" \
	GPGKEY_USER="prvt CI <ci@prvt>" \
	  .bin/go-acc $(shell go list ./... | grep -v prvt/wasm) -- -v -ldflags "-X github.com/ItalyPaleAle/prvt/buildinfo.Production=1" -short
	# Remove generated (.pb.go ones) files from coverage report
	cat coverage.txt| grep -v ".pb.go:" > coverage-filtered.txt
