.PHONY: test test-verbose coverage race build build-pi run \
	deploy deploy-frontend deploy-daemon \
	deploy-snespi deploy-milkpi

# Which box is being deployed. Its settings live in deploy/environments/<name>.
# Leaving it empty builds the frontend with its defaults and installs the
# daemon from the generic template, which is what the targets did before the
# environments existed.
KIOSK_ENV ?=

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

deploy-snespi:
	$(MAKE) deploy KIOSK_ENV=snespi

deploy-milkpi:
	$(MAKE) deploy KIOSK_ENV=milkpi

deploy: deploy-frontend deploy-daemon

deploy-frontend:
ifeq ($(KIOSK_ENV),)
	cd frontend && pnpm build
else
	./deploy/frontend/build.sh $(KIOSK_ENV)
endif
	./deploy/frontend/install.sh

deploy-daemon:
	KIOSK_ENV=$(KIOSK_ENV) ./deploy/daemon/install.sh