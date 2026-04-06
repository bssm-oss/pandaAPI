BIN := pandaapi

.PHONY: build clean deps test

build:
	go build -o $(BIN) ./...

clean:
	rm -f $(BIN)

deps:
	go mod tidy

test:
	go test ./...
