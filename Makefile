BIN=bin
BIN_NAME=gopicto

PKG_NAME = github.com/nmaupu/gopicto
TAG_NAME ?= $(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD)
LDFLAGS = -ldflags="-X '$(PKG_NAME)/cli.ApplicationVersion=$(TAG_NAME)' -X '$(PKG_NAME)/cli.BuildDate=$(shell date)'"

all: $(BIN_NAME)

$(BIN_NAME): $(BIN)
	go build -o $(BIN)/$(BIN_NAME) $(LDFLAGS)

.PHONY: release
release:
	GOOS=linux  GOARCH=amd64 go build -o $(BIN)/$(BIN_NAME)-$(TAG_NAME)-linux_x64    $(LDFLAGS)
	GOOS=darwin GOARCH=amd64 go build -o $(BIN)/$(BIN_NAME)-$(TAG_NAME)-darwin_x64   $(LDFLAGS)
	GOOS=darwin GOARCH=arm64 go build -o $(BIN)/$(BIN_NAME)-$(TAG_NAME)-darwin_arm64 $(LDFLAGS)
	@echo
	@echo Changelog:
	$(eval from:=$(shell git tag | sort | tail -2 | head -1))
	$(eval to:=$(shell git tag | sort | tail -1))
	@git log --pretty="- %C(auto)%h %s" "$(from)..$(to)" | cat

$(BIN):
	mkdir -p $(BIN)

.PHONY: clean
clean:
	rm -rf $(BIN)
