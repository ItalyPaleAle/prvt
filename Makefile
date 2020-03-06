all: build

clean:
	rm -fv ui/dist/*

build: build-ui
	packr2
	go build -o bin

build-ui:
	(cd ui; npm ci; APP_VERSION="0.1.0" npm run build)
