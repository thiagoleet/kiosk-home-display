.PHONY: test test-verbose coverage race build run

test:
	cd backend && gotestsum ./...

test-verbose:
	cd backend && go test -v ./...

coverage:
	cd backend && go test ./... -coverprofile=coverage.out
	cd backend && go tool cover -html=coverage.out

race:
	cd backend && go test -race ./...

build:
	cd backend && go build -o kiosk ./cmd/kiosk

run:
	cd backend && go run ./cmd/kiosk