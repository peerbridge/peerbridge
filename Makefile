#!make

GOOS:=$(shell go env GOOS)
GOARCH:=$(shell go env GOARCH)

docs:
	@godoc -http=:6060

build:
	@GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=1 go build -a -o "bin/peerbridge-${GOOS}-${GOARCH}" -ldflags="-extldflags '-static' -s -w"

bin: bin-linux bin-windows

bin-local:
	@docker build --build-arg GOOS=${GOOS} --build-arg GOARCH=${GOARCH} --target bin --output bin/ --platform local  .

bin-linux:
	@docker build --platform linux/amd64 --build-arg GOOS=linux --build-arg GOARCH=amd64 --build-arg CC=gcc --build-arg CXX=g++ --target bin --output bin/ .

bin-darwin:
	@echo "Darwin cross compilation using docker is currently not supported."

bin-windows:
	@docker build --platform windows/amd64 --build-arg GOOS=windows --build-arg GOARCH=amd64 --build-arg CC=x86_64-w64-mingw32-gcc --build-arg CXX=x86_64-w64-mingw32-g++ --target bin --output bin/ .

fmt:
	@gofmt -w .

test:
	@go test -v ./...

coverage:
	@go test ./... -cover -coverprofile=c.out
	@go tool cover -html=c.out -o coverage.html
