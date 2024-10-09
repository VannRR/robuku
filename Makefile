.DEFAULT_GOAL := build
.PHONY: fmt vet build run install clean test

APP_NAME := robuku
INSTALL_DIR := ~/.config/rofi/scripts/

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -ldflags="-w -s" -o $(APP_NAME)

run: build
	rofi -show $(APP_NAME) -mode $(APP_NAME) -modi $(APP_NAME)':'$(APP_NAME)

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(APP_NAME) $(INSTALL_DIR)

clean:
	go clean
	rm -f $(APP_NAME)

test:
	go test ./...
