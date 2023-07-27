install:
	go install -v

build:
	go build -race ./...

lint:
	go vet ./...
	revive ./...

test:
	go test -count=1 -race -v ./... --cover

benchmark:
	go test -count=1 -bench=. -run "Benchmark"

integration-test:
	go run -C ./tests -race .

deps:
	go install github.com/mgechev/revive@latest

clean:
	go clean
