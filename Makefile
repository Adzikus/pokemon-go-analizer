.PHONY: build test build-test clean

build:
	cd src/pokemon && go build -o ../../bin/pokemon .

test:
	cd src/pokemon && go test ./...

build-test: build test

clean:
	rm -f bin/pokemon
