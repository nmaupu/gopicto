BIN=bin
BIN_NAME=gopicto

all: $(BIN_NAME)

$(BIN_NAME): $(BIN)
	go build -o $(BIN)/$(BIN_NAME)

$(BIN):
	mkdir -p $(BIN)

.PHONY: clean
clean:
	rm -rf $(BIN)
