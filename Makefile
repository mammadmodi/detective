IMAGE_FLAG := $(shell git rev-parse --abbrev-ref HEAD | tr '/' '-') #get image flag from current branch name.
IMAGE_TAG  ?= detective:${IMAGE_FLAG}

dependencies:
	go mod vendor

unit-test: dependencies
	go test -mod=vendor --race -v -coverprofile .coverage.out ./...
	go tool cover -func .coverage.out

compile: dependencies
	go build -mod=vendor --ldflags "-X main.CommitRefName=$(COMMIT_REF_SLUG) -X main.CommitSHA=$(COMMIT_SHORT_SHA) -X main.BuildDate=$(CURRENT_DATETIME) -linkmode external -extldflags '-static'" -o detective-server ./cmd/server/main.go

build-image:
	docker build -f ./build/Dockerfile -t ${IMAGE_TAG} .

up: down build-image
	docker-compose up -d

down:
	docker-compose down
