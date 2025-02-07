BINARY_NAME=main.out

.PHONY: all test clean

build:
	go build

run:
	./$(BINARY_NAME)

dev:
	gow run .