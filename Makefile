GOCMD=go
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod

all:  build-coordinator build-scheduler build-worker build-cli

build-coordinator:
	$(GOBUILD) -v -o dist/archer-coordinator cmd/coordinator/main.go
build-scheduler:
	$(GOBUILD) -v -o dist/archer-scheduler cmd/scheduler/main.go
build-worker:
	$(GOBUILD) -v -o dist/archer-worker cmd/worker/main.go
build-cli:
	$(GOBUILD) -v -o dist/archer-cli cmd/cli/main.go
tidy:
	$(GOMOD) tidy