APP_VERSION ?= canary

all: build

clean:
	rm -rfv ui/dist/*
	rm -rfv .bin bin
	packr2 clean

build: build-ui
	packr2
	go build -o bin

build-ui:
	(cd ui; npm ci; APP_VERSION="$(APP_VERSION)" npm run build)
