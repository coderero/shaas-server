name: IOT-Server
version: 0.1.0

RED=\033[0;31m
GREEN=\033[0;32m
BLUE=\033[0;34m
NC=\033[0m # No Color


build:
	@go build ./cmd/main.go bin/iot-server

run:
	@echo "${GREEN}Running the server...${NC}"
	@./bin/iot-server
	@echo "${GREEN}Server is running!${NC}"
	@echo "${GREEN}Press Ctrl+C to stop the server.${NC}"

