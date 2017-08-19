all: build

GOOS?=linux
REGISTRY?=stor.highloadcup.ru/travels/shy_caracal
APP?=travels

build: 
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -installsuffix cgo \
		-o ./bin/${APP} ./cmd/server

container: build
	docker build -t ${APP}:latest .

run: container
	docker run --rm -p 80:80 -t ${APP}:latest
