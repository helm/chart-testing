

.PHONY: build
build:
	go build -o ct app/main.go

.PHONY: test
test:
	go test ./...

.PHONY: release
release: test
	./tag.sh
