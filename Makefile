all: build

# Local: DOCKERFILE=Dockerfile-local make run

GOOS?=linux
REGISTRY?=stor.highloadcup.ru/travels/shy_caracal
APP?=travels

build: 
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -installsuffix cgo \
		-o ./bin/${APP} ./cmd/server

local: build
	docker build -t ${APP}:latest -f Dockerfile-local .

run: local
	docker stop ${APP} || true 
	docker run --name ${APP} --rm -p 80:80 -t ${APP}:latest

container: build
	docker build -t ${APP}:latest -f Dockerfile .

deploy: container
	docker tag ${APP}:latest ${REGISTRY}
	docker push ${REGISTRY}
