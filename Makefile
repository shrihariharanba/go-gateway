# Makefile

# Variables
APP_NAME := go-gateway
SRC := cmd/gateway/main.go

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME) $(SRC)

# Build for Windows
build-windows:
	go build -o $(APP_NAME).exe $(SRC)

# Run on Linux
run-linux: build-linux
	./$(APP_NAME)

# Run on Windows
run-windows: build-windows
	$(APP_NAME).exe
