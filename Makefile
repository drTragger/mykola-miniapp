APP_NAME=mykola-miniapp
BUILD_DIR=bin
BINARY=$(BUILD_DIR)/$(APP_NAME)
CMD_PATH=./cmd/mykola-miniapp
SERVICE_NAME=mykola-miniapp

.PHONY: help tidy fmt build run stop restart status logs pull deploy clean

help:
	@echo "Available commands:"
	@echo "  make tidy     - run go mod tidy"
	@echo "  make fmt      - format Go code"
	@echo "  make build    - build binary"
	@echo "  make run      - build and run locally"
	@echo "  make pull     - git pull"
	@echo "  make restart  - restart systemd service"
	@echo "  make status   - show systemd service status"
	@echo "  make logs     - show service logs"
	@echo "  make deploy   - pull, tidy, build, restart"
	@echo "  make clean    - remove build artifacts"

tidy:
	go mod tidy

fmt:
	go fmt ./...

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BINARY) $(CMD_PATH)

run: build
	BOT_TOKEN=$$BOT_TOKEN APP_ADDR=:8090 $(BINARY)

pull:
	git pull

restart:
	sudo systemctl restart $(SERVICE_NAME)

status:
	sudo systemctl status $(SERVICE_NAME)

logs:
	sudo journalctl -u $(SERVICE_NAME) -f

deploy: pull tidy build restart status

clean:
	rm -rf $(BUILD_DIR)