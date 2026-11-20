.PHONY: build run test docker-build docker-push lint tidy

APP_NAME=ag-ecommerce-backend
IMAGE=ghcr.io/ahyan-1986/$(APP_NAME)
TAG?=latest

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./... -v -race

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-push: docker-build
	docker push $(IMAGE):$(TAG)
