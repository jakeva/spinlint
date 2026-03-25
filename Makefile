.PHONY: build test vet lint clean

BINARY  := spinlint
BIN_DIR := bin

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

build: $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/spinlint

test:
	go test -v ./...

vet:
	go vet ./...

lint: vet
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR)
