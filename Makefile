VERSION := $(shell git rev-parse --abbrev-ref HEAD | tr '/' '-') #get image flag from current branch name.
COMMIT_SHA = $(shell git rev-parse HEAD)
BUILD_DATE = $(shell date -u +%Y/%m/%d-%H/%M/%S)
IMAGE_TAG  ?= detective:${VERSION}

dependencies:
	go mod vendor

unit-test: dependencies
	go test -mod=vendor --race -v -coverprofile .coverage.out ./...
	go tool cover -func .coverage.out

compile: dependencies
	go build -mod=vendor \
	 --ldflags "-X main.CommitRefName=$(VERSION) -X main.CommitSHA=$(COMMIT_SHA) -X main.BuildDate=$(BUILD_DATE) -linkmode external -extldflags '-static'" \
	 -o detective-server ./cmd/server/main.go

build-image:
	docker build -f ./build/Dockerfile -t ${IMAGE_TAG} .

up: down
	docker-compose up -d

down:
	docker-compose down
