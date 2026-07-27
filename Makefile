BINARY := sqlcleaner

.PHONY: all build test vet clean

all: test build

build:
	go build -o $(BINARY) ./cmd/sqlcleaner

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
