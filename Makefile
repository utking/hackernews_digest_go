build:
	go build -o bin/hn_digest -ldflags="-s -w -extldflags=-static" main.go

build-all: build

tar: build
	tar --directory=. --transform='s|bin||' -czvf build.tgz bin/hn_digest config.example.json

clean:
	rm -vf bin/hn_digest *.tgz