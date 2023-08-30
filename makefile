APP_NAME=dts-cops-pg-mongo
DEFAULT_PORT=8200
POSTGRES_CONTAINER?=postgres
TOOLS_IMAGE=postgres:latest
APP_ENVIRONMENT=docker run --rm -v ${PWD}:/${APP_NAME} -w /${APP_NAME} --net=host ${TOOLS_IMAGE}

.PHONY: setup init build dev test migrate-up migrate-down ci

setup:
	docker pull ${TOOLS_IMAGE}

init:
	@if [[ "$(docker images -q sp/dts-cops-pg-mongo:latest 2> /dev/null)" == "" ]]; then \
		docker pull ${TOOLS_IMAGE}; \
	fi

	make remove-infras
	docker-compose up -d
	@echo "Waiting for database connection..."
	@while ! docker exec ${POSTGRES_CONTAINER} pg_isready > /dev/null; do \
		sleep 1; \
	done

remove-infras:
	docker-compose down --remove-orphans --volumes

build:
	env GOOS=darwin GOARCH=amd64 go build -o bin ./...

dev:
	go run ./cmd/server/main.go

generate:
	go run ./cmd/generator/main.go
