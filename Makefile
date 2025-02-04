BINARY_NAME=main.out

.PHONY: all test clean

run:
	go build
	go auth

test:
	go test ./tests/..._test -v

coverage:
	go test -coverprofile=coverage.out ./internal/presentation/controllers/... ./internal/data/usecase/... ./internal/infra/db/mongodb/... ./internal/utils/... ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	xdg-open coverage.html