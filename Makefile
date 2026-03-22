APP_NAME=mykola-miniapp
BUILD_DIR=bin
BINARY=$(BUILD_DIR)/$(APP_NAME)
CMD_PATH=./cmd/mykola-miniapp
SERVICE_NAME=mykola-miniapp
FRONTEND_DIR=frontend

.PHONY: help tidy fmt frontend-install frontend-build build run pull restart reload status logs deploy clean

help:
	@echo "make tidy             - go mod tidy"
	@echo "make fmt              - go fmt ./..."
	@echo "make frontend-install - npm install in frontend"
	@echo "make frontend-build   - build Vue frontend"
	@echo "make build            - install frontend deps + build frontend + build Go app"
	@echo "make run              - run app locally"
	@echo "make pull             - git pull"
	@echo "make restart          - restart systemd service"
	@echo "make reload           - daemon-reload + restart"
	@echo "make status           - show service status"
	@echo "make logs             - tail service logs"
	@echo "make deploy           - pull + tidy + build + restart + status"
	@echo "make clean            - delete build artifacts"

tidy:
	go mod tidy

fmt:
	go fmt ./...

frontend-install:
	cd $(FRONTEND_DIR) && npm install

frontend-build:
	cd $(FRONTEND_DIR) && npm run build

build: frontend-install frontend-build
	mkdir -p $(BUILD_DIR)
	go build -o $(BINARY) $(CMD_PATH)

run: build
	BOT_TOKEN=$$BOT_TOKEN APP_ADDR=:8090 $(BINARY)

pull:
	git pull

restart:
	sudo systemctl restart $(SERVICE_NAME)

reload:
	sudo systemctl daemon-reload
	sudo systemctl restart $(SERVICE_NAME)

status:
	sudo systemctl status $(SERVICE_NAME) --no-pager

logs:
	sudo journalctl -u $(SERVICE_NAME) -f

deploy: pull tidy build restart status

clean:
	rm -rf $(BUILD_DIR)
	rm -rf internal/web/dist