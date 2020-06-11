#!make

container-build:
	@docker-compose -f deployments/docker-compose.yml build

container-start:
	@docker-compose -f deployments/docker-compose.yml up -d

container-stop:
	@docker-compose -f deployments/docker-compose.yml down
