BIN=bin
BIN_NAME=gopicto

PKG_NAME = github.com/nmaupu/gopicto
TAG_NAME = $(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD)
LDFLAGS = -ldflags="-X '$(PKG_NAME)/cli.ApplicationVersion=$(TAG_NAME)' -X '$(PKG_NAME)/cli.BuildDate=$(shell date)'"

all: $(BIN_NAME)

$(BIN_NAME): $(BIN)
	go build -o $(BIN)/$(BIN_NAME) $(LDFLAGS)

$(BIN):
	mkdir -p $(BIN)

.PHONY: clean
clean:
	rm -rf $(BIN)
