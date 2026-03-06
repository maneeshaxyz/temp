.DEFAULT_GOAL := run

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./... -cover -race

lint: fmt vet test
	@echo "Code passed fmt, vet, and tests"

gci:
	golangci-lint run

#GOEXPERIMENT=jsonv2
build: lint
	go build -o ./bin/intake ./cmd/intake


dbuild:
	docker build -t intake .

drun: dbuild
	docker run --rm --name intake \
		-p 8080:8080 \
		intake

linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/intake ./cmd/intake

run: build
	./bin/intake

clean:
	rm -f ./bin/intake

# --- Phony targets ---
.PHONY: fmt vet test lint gci build dbuild drun linux run clean
