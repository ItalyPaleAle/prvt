APP_VERSION ?= canary

# Performs a builld
all: build

# Fetches the tools that are required to build prvt
get-tools:
	mkdir -p .bin
	curl -sf https://gobinaries.com/github.com/ory/go-acc@v0.2.6 | PREFIX=.bin/ sh
	curl -sf https://gobinaries.com/github.com/gobuffalo/packr/packr2@v2.7.1 | PREFIX=.bin/ sh

clean:
	rm -rfv ui/dist/*
	rm -rfv .bin/prvt* bin
	.bin/packr2 clean

build: build-ui
	.bin/packr2
	go build -o bin

build-ui:
	(cd ui; npm ci; APP_VERSION="$(APP_VERSION)" npm run build)

test:
	GPGKEY_ID="0x4C6D7DB1D92F58EE"
	GPGKEY_USER="prvt CI <ci@prvt>"
	.bin/go-acc ./... -- -v
