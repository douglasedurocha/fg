VERSION := 0.1.0
COMMIT := $(shell git rev-parse --short HEAD || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/fhir-guard/fg/cmd.Version=$(VERSION) -X github.com/fhir-guard/fg/cmd.Commit=$(COMMIT) -X github.com/fhir-guard/fg/cmd.BuildDate=$(BUILD_DATE)

.PHONY: build clean test test-unit test-integration test-coverage test-all run

# Compilação padrão
build:
	go build -ldflags "$(LDFLAGS)" -o bin/fg main.go

# Compilação para desenvolvimento rápido
dev:
	go build -o bin/fg main.go

# Limpeza de artefatos de compilação
clean:
	rm -rf bin/
	rm -f coverage.out

# Execução de testes unitários
test-unit:
	go test -v ./cmd/... ./config/... -short

# Execução de testes de integração
test-integration:
	go test -v ./... -run Integration -count=1

# Execução de todos os testes
test-all: test-unit test-integration

# Execução de testes com cobertura
test-coverage:
	go test -v ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Compatibilidade com o comando test antigo
test: test-unit

# Execução da aplicação
run: dev
	./bin/fg

# Instalação da aplicação no sistema
install: build
	go install -ldflags "$(LDFLAGS)"

# Construir binários para várias plataformas (cross-compile)
build-all:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/fg-windows-amd64.exe main.go
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/fg-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/fg-darwin-amd64 main.go

# Execução da aplicação com informações de versão
version:
	@echo "FHIR Guard CLI v$(VERSION) ($(COMMIT)) built at $(BUILD_DATE)"

# Verifica os testes sem realmente executá-los
test-dry-run:
	go test -v ./... -run=TestSuite -count=1 -test.list=.

# Executa os testes com o race detector
test-race:
	go test -v -race ./...

# Linter para código Go
lint:
	go vet ./...
	@if command -v staticcheck > /dev/null; then \
		staticcheck ./...; \
	else \
		echo "Staticcheck não encontrado. Execute 'go install honnef.co/go/tools/cmd/staticcheck@latest' para instalá-lo."; \
	fi

# Verificação completa: linting, testes unitários, testes de integração e cobertura
check: lint test-unit test-coverage 

# Executar um teste específico (útil para depuração e ambientes Windows)
test-single:
	@echo "Executando teste $(TEST)"
	go test -v ./$(PKG)/... -run="$(TEST)" -timeout 120s 