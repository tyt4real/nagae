.PHONY: generate build run dev clean

generate:
	templ generate

build: generate
	go build -o tmp/webmail ./cmd/server

run: generate
	go run ./cmd/server

dev:
	air

clean:
	rm -rf bin/
	find . -name '*_templ.go' -delete
