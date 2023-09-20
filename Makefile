BIN_BINARY_NAME=docker-events-monitor
BIN_SUPERVISOR=supervisor-events-listener
BIN_SYSTEMD=systemd-events-listener

default:
	go build -o $(BIN_BINARY_NAME) cmd/main.go
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run ./$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"

sup:
	go build -o $(BIN_SUPERVISOR) cmd/supervisor-events-listener/main.go
	@echo "Build $(BIN_SUPERVISOR) successfully. You can run ./$(BIN_SUPERVISOR) now.If you can't see it soon,wait some seconds"

systemd:
	go build -o $(BIN_SYSTEMD) cmd/systemd-events-listener/main.go
	@echo "Build $(BIN_SYSTEMD) successfully. You can run ./$(BIN_SYSTEMD) now.If you can't see it soon,wait some seconds"

update:
	go mod tidy
	go mod vendor