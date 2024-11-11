VERSION:=commit_$(shell git rev-parse --short HEAD)_time_$(shell date +"%Y-%m-%dT%H:%M:%S")
BUILDTIME := $(shell date +"%Y-%m-%dT%H:%M:%S")

GOLDFLAGS += -X main.Version=$(VERSION)
GOLDFLAGS += -X main.Buildtime=$(BUILDTIME)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

linux-amd64-api:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o quai-sync-linux-amd64 $(GOFLAGS) ./cmd/main.go

linux-arm64-api:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o quai-sync-linux-arm64 $(GOFLAGS) ./cmd/main.go

darwin-arm64-api:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o quai-sync-darwin-arm64 $(GOFLAGS) ./cmd/main.go

darwin-amd64-api:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o quai-sync-darwin-amd64 $(GOFLAGS) ./cmd/main.go

windows-amd64-api:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o quai-sync-windows-amd64 $(GOFLAGS) ./cmd/main.go

build:
	go build $(GOFLAGS) -o quai-sync ./cmd/main.go