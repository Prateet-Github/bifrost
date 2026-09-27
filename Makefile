BINARY=bifrost

.PHONY: all
all: fmt test

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: race
race:
	go test -race ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem ./...

.PHONY: build
build:
	go build -o bin/$(BINARY) ./cmd/bifrost

.PHONY: clean
clean:
	rm -rf bin