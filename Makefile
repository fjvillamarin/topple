VERSION := 0.1.0

.PHONY: run
run:
	go run ./cmd/... compile examples/ -r -d

.PHONY: watch
watch:
	go run ./cmd/... watch examples/ -d

.PHONY: scan
scan:
	go run ./cmd/... scan examples/python/ -d -w -r

.PHONY: parse
parse:
	go run ./cmd/... parse examples/python/01_hello_world/01_hello_world.py -d -w -r

build:
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/biscuit ./cmd/...

.PHONY: fmt	
fmt:
	go fmt ./...