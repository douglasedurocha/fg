VERSION := 0.1.0
COMMIT := $(shell git rev-parse --short HEAD || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/fhir-guard/fg/cmd.Version=$(VERSION) -X github.com/fhir-guard/fg/cmd.Commit=$(COMMIT) -X github.com/fhir-guard/fg/cmd.BuildDate=$(BUILD_DATE)

.PHONY: build clean test run

# Compilação padrão
build:
	go build -ldflags "$(LDFLAGS)" -o bin/fg main.go

# Compilação para desenvolvimento rápido
dev:
	go build -o bin/fg main.go

# Limpeza de artefatos de compilação
clean:
	rm -rf bin/

# Execução de testes
test:
	go test -v ./...

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