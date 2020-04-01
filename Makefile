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
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...
