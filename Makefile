.PHONY: build
build:
	go build cmd/uptimerobot-tooling/uptimerobot-tooling.go

.PHONY: test
test:
	go test -v ./...