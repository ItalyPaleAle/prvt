APP_VERSION ?= dev
BUILD_TIME := $(shell date -u +'%Y-%m-%dT%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD)

# Performs a builld
all: build

# Fetches the tools that are required to build prvt
get-tools:
	mkdir -p .bin
	curl -sf https://gobinaries.com/github.com/ory/go-acc@v0.2.6 | PREFIX=.bin/ sh
	curl -sf https://gobinaries.com/github.com/markbates/pkger/cmd/pkger@v0.17.1 | PREFIX=.bin/ sh

# Clean all compiled files
clean:
	rm -rfv ui/dist/* || true
	rm -rfv .bin/prvt* bin || true
	rm -v pkged.go || true

# Build the entire app
build: build-ui build-app

# Build the Go code
build-app:
	.bin/pkger list
	.bin/pkger
	go build \
	  -ldflags "-X github.com/ItalyPaleAle/prvt/buildinfo.Production=1 -X github.com/ItalyPaleAle/prvt/buildinfo.AppVersion=$(APP_VERSION) -X github.com/ItalyPaleAle/prvt/buildinfo.BuildID=$(APP_VERSION) -X github.com/ItalyPaleAle/prvt/buildinfo.BuildTime=$(BUILD_TIME) -X github.com/ItalyPaleAle/prvt/buildinfo.CommitHash=$(COMMIT_HASH)" \
	  -o bin

# Buold the web UI
build-ui:
	(cd ui; npm ci; APP_VERSION="$(APP_VERSION)" npm run build)

# Run tests
test:
	GPGKEY_ID="0x4C6D7DB1D92F58EE"
	GPGKEY_USER="prvt CI <ci@prvt>"
	.bin/go-acc ./... -- -v
