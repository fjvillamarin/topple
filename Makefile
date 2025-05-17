VERSION := 0.1.0

.PHONY: run
run:
	go run ./cmd/... compile examples/ -r -d

.PHONY: watch
watch:
	go run ./cmd/... watch examples/ -d

.PHONY: tests
test:
	go test -v ./...

build:
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/biscuit ./cmd/...

.PHONY: fmt	
fmt:
	go fmt ./...