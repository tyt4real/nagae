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
docker:
	docker build -t webmail .
 
docker-run:
	docker run --rm -p 8080:8080 -v $(PWD)/config.json:/app/config.json:ro webmail
