GOBIN             := $(shell pwd)/bin

.PHONY: all

all:
	mkdir -p $(GOBIN)
	go build -mod=vendor -o $(GOBIN) ./...

run:
	./scripts/run.sh

dc-down:
	docker-compose down --remove-orphans

dc-up:
	docker-compose up -d --build webserver
