.PHONY: test test-verbose coverage build run

test:
	cd backend && gotestsum ./...

test-verbose:
	cd backend && go test -v ./...

coverage:
	cd backend && go test ./... -coverprofile=coverage.out
	cd backend && go tool cover -html=coverage.out

build:
	cd backend && go build -o kiosk ./cmd/kiosk

run:
	cd backend && go run ./cmd/kiosk