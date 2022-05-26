.PHONY: build test

build:
	go build -o conduit-connector-zendesk cmd/main.go

test:
	go test $(GOTEST_FLAGS) -race ./...

