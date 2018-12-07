.PHONY: setup

all: setup test

setup:
	@go get -v -t $(SRC)
	@go get -v golang.org/x/lint/golint

test:
	@go test --race --covermode=atomic --coverprofile=coverage.txt ./...
