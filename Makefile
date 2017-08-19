all: build

# Local: DOCKERFILE=Dockerfile-local make run

GOOS?=linux
REGISTRY?=stor.highloadcup.ru/travels/shy_caracal
APP?=travels
DOCKERFILE?=Dockerfile

build: 
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -installsuffix cgo \
		-o ./bin/${APP} ./cmd/server

container: build
	docker build -t ${APP}:latest -f ${DOCKERFILE} .

run: container
	docker stop ${APP} || true 
	docker run --name ${APP} --rm -p 80:80 -t ${APP}:latest

deploy: container
	docker tag ${APP}:latest ${REGISTRY}
	docker push ${REGISTRY}
