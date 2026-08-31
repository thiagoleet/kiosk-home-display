.PHONY: test test-verbose coverage race build build-pi run

test:
	cd daemon && gotestsum ./...

test-verbose:
	cd daemon && go test -v ./...

coverage:
	cd daemon && go test ./... -coverprofile=coverage.out
	cd daemon && go tool cover -html=coverage.out

race:
	cd daemon && go test -race ./...

build:
	cd daemon && go build -o kiosk ./cmd/kiosk

build-pi:
	cd daemon && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o kiosk ./cmd/kiosk

run:
	cd daemon && go run ./cmd/kiosk