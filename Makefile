.PHONY: run
run:
	go run cmd/gophermart/main.go

.PHONY: fmt
fmt:
	goimports -local "github.com/paramonies/ya-gophermart" -w cmd internal pkg/lg pkg/queue

.PHONY: lint
lint:
	golangci-lint run -v ./...