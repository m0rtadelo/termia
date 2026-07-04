BINARY := termia
PKG := ./...

.PHONY: build run test vet fmt lint clean install

build:
	go build -o $(BINARY) .

run:
	go run .

test:
	go test $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -w .

lint: fmt vet

clean:
	rm -f $(BINARY)

install:
	go install .
