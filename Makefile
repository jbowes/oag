check: lint cover build
ci: lint report build

build:
	go build .

test:
	go test -v ./...

cover:
	go test -cover ./...

report:
	@echo "" > coverage.txt
	@for d in $$(go list ./...); do \
		go test -race -coverprofile=profile.out -covermode=atomic $$d; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi \
	done

lint:
	gometalinter --fast --disable=gocyclo --disable=gosec --enable=gofmt --vendor ./...
