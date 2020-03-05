all: build

clean:
	rm -fv ui/dist/*.css ui/dist/*.css.map ui/dist/*.js ui/dist/*.js.map

build: build-ui
	packr2
	go build -o bin

build-ui:
	(cd ui; npm ci; npm run build)
