check: lint cover build

build:
	go build .

cover:
	go test -cover $$(go list ./... | grep -v vendor)

lint:
	gometalinter --fast --disable=gocyclo --vendor ./...
