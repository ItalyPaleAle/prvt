APP_VERSION ?= canary

all: build

clean:
	rm -rfv ui/dist/* || true
	rm -v pkged.go || true

build: build-ui
	pkger list
	pkger
	go build -o bin

build-ui:
	(cd ui; npm ci; APP_VERSION="$(APP_VERSION)" npm run build)
