.PHONY: build run test docker-up docker-down clean archive

build:
	go build -o bin/server-inventory .

run:
	go run .

test:
	go test -v ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

clean:
	rm -rf bin/ solution.zip

archive: clean
	zip -r solution.zip . -x "*.git*" "*.DS_Store*" "bin/*" "solution.zip"
