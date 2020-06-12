#!make

container-build:
	@docker-compose -f deployments/docker-compose.yml build

container-start:
	@docker-compose -f deployments/docker-compose.yml up -d

container-stop:
	@docker-compose -f deployments/docker-compose.yml down

docs:
	@godoc -http=:6060

build:
	@docker build --file deployments/Dockerfile --target bin --output bin/ --platform local .

build-windows:
	@docker build --file deployments/Dockerfile --target bin --output bin/ --platform windows/amd64 .
