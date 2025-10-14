VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null | sed -E 's/(v[0-9]+\.[0-9]+\.[0-9]+).*/\1/' || git rev-parse --short HEAD)
PKG_LIST := $(shell go list ./... > /tmp/packages.txt && python3 ./scripts/coverage_packages.py)

################################################################################
## Dependencies
################################################################################
.PHONY: all generate/mocks init

init:
	@echo "----------------------------------------------------------------"
	@echo " 📦 Installing necessary binaries"
	@echo "----------------------------------------------------------------"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5
	go install github.com/matm/gocov-html/cmd/gocov-html@v1.4.0
	go install github.com/axw/gocov/gocov@v1.2.0
	go install github.com/cespare/reflex@v0.3.1
	go install github.com/vektra/mockery/v3@v3.5.0

generate: generate/mocks

generate/mocks:
	@echo "----------------------------------------------------------------"
	@echo " 🔌 Generating mocks"
	@echo "----------------------------------------------------------------"
	mockery

all: tidy vet lint coverage

################################################################################
## Common commands
################################################################################

tidy:
	@echo "----------------------------------------------------------------"
	@echo " 🗑️ Cleaning code..."
	@echo "----------------------------------------------------------------"
	go mod tidy
	go fmt ./...

vet:
	@echo "----------------------------------------------------------------"
	@echo " 🔍 Vetting code..."
	@echo "----------------------------------------------------------------"
	go vet ./...

lint:
	@echo "----------------------------------------------------------------"
	@echo " 🔍 Linting code..."
	@echo "----------------------------------------------------------------"
	golangci-lint run --timeout=10m ./...


################################################################################
## Testing commands
################################################################################

test:
	@echo "----------------------------------------------------------------"
	@echo " ✅ Testing..."
	@echo "----------------------------------------------------------------"
	go test -tags test ./...

test-no-cache:
	@echo "----------------------------------------------------------------"
	@echo " ✅ Testing..."
	@echo "----------------------------------------------------------------"
	go test -tags test ./... -count=1

coverage:
	@echo "----------------------------------------------------------------"
	@echo " ✅ Testing with coverage..."
	@echo "----------------------------------------------------------------"
	go test -tags test -count=1 -coverpkg="$(PKG_LIST)" ./... -coverprofile=coverage.out
	@test_result=$$?; \
	if [ $$test_result -ne 0 ]; then \
			exit $$test_result; \
	fi; \
	gocov convert coverage.out | gocov report

coverage/html:
	@echo "----------------------------------------------------------------"
	@echo " ✅ Testing with coverage (HTML report)..."
	@echo "----------------------------------------------------------------"
	go test -tags test -count=1 -coverpkg="$(PKG_LIST)" ./... -coverprofile=coverage.out
	@test_result=$$?; \
	if [ $$test_result -ne 0 ]; then \
			exit $$test_result; \
	fi; \
	gocov convert coverage.out | gocov-html > coverage.html
