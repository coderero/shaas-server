name: IOT-Server
version: 0.1.0

RED=\033[0;31m
GREEN=\033[0;32m
BLUE=\033[0;34m
NC=\033[0m # No Color


build:
	@go build -o bin/iot-server ./cmd/main.go

run:
	@echo "${GREEN}Running the server...${NC}"
	@./bin/iot-server serve
	@echo "${GREEN}Server is running!${NC}"
	@echo "${GREEN}Press Ctrl+C to stop the server.${NC}"

