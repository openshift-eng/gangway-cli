all: test build

verify: lint

build:
	go build .

test:
	go test ./...

mapping: build

lint:
	./hack/go-lint.sh run ./...

clean:
	rm -f gangway-cli
