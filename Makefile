install:
	go install -v

build:
	go build -v ./...

lint:
	go vet ./...
	revive ./...

test:
	go test -v ./... --cover

deps:
	go install github.com/mgechev/revive@latest

clean:
	go clean
