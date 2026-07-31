.PHONY: build run dev clean

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

dev:
	go run ./cmd/server

clean:
	rm -rf bin/

docker-build:
	docker build -t devplatform/api:latest .

docker-run:
	docker-compose up -d

test:
	go test ./...
