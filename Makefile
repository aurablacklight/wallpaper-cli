.PHONY: build test clean fmt vet install lint

BINARY_NAME=wallpaper-cli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS=-ldflags "-X github.com/user/wallpaper-cli/cmd.version=$(VERSION) \
                  -X github.com/user/wallpaper-cli/cmd.commit=$(COMMIT) \
                  -X github.com/user/wallpaper-cli/cmd.date=$(DATE) \
                  -s -w"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet
	golangci-lint run || true

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/ || cp $(BINARY_NAME) ~/go/bin/

# Cross-compilation
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
