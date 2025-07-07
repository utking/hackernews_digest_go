build:
	go build -o bin/hn_digest -ldflags="-s -w -extldflags=-static" main.go

build-all: build

tar: build
	tar --directory=. --transform='s|bin||' -czvf build.tgz bin/hn_digest config.example.json

clean:
	rm -vf bin/hn_digest *.tgz

lint:
	revive ./...

install-lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

lintci:
	golangci-lint run

update:
	go get -u all

install-revive:
	go install github.com/mgechev/revive@latest

test:
	go test -cover -v ./...
