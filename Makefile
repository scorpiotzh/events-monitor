# build file
GOCMD=go
# Use -a flag to prevent code cache problems.
GOBUILD=$(GOCMD) build -ldflags -s -v -i

BIN_BINARY_NAME=docker-events-monitor
docker:
	$(GOBUILD) -o $(BIN_BINARY_NAME) cmd/main.go
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run ./$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"

BIN_SUPERVISOR=supervisor-events-listener
sup:
	$(GOBUILD) -o $(BIN_SUPERVISOR) cmd/supervisor-events-listener/main.go
	@echo "Build $(BIN_SUPERVISOR) successfully. You can run ./$(BIN_SUPERVISOR) now.If you can't see it soon,wait some seconds"

BIN_SYSTEMD=systemd-events-listener
systemd:
	$(GOBUILD) -o $(BIN_SYSTEMD) cmd/systemd-events-listener/main.go
	@echo "Build $(BIN_SYSTEMD) successfully. You can run ./$(BIN_SYSTEMD) now.If you can't see it soon,wait some seconds"

update:
	go mod tidy
	go mod vendor