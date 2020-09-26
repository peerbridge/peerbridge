#!make

docs:
	@godoc -http=:6060

build:
	@docker build --file deployments/Dockerfile --target bin --output bin/ --platform local .

build-windows:
	@docker build --file deployments/Dockerfile --target bin --output bin/ --platform windows/amd64 .

fmt:
	@gofmt -w .

test:
	@go test -v ./...

coverage:
	@go test ./... -cover -coverprofile=c.out
	@go tool cover -html=c.out -o coverage.html
