BIN=bin
BIN_NAME=gopicto

PKG_NAME = github.com/nmaupu/gopicto
LDFLAGS = -ldflags="-X '$(PKG_NAME)/cli.ApplicationVersion=$(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)' -X '$(PKG_NAME)/cli.BuildDate=$(shell date)'"

all: $(BIN_NAME)

$(BIN_NAME): $(BIN)
	go build -o $(BIN)/$(BIN_NAME) $(LDFLAGS)

$(BIN):
	mkdir -p $(BIN)

.PHONY: clean
clean:
	rm -rf $(BIN)
