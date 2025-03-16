APP_NAME = connect6
GO = go
GO_FLAGS = -v
GO_BUILD = $(GO) build $(GO_FLAGS)
GOOS = linux
GOARCH = amd64

.PHONY: all build-linux clean

all: build-linux

build-linux:
	@echo "Compilando para Linux..."
	@mkdir -p bin
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_BUILD) -o bin/$(APP_NAME) .
	@echo "Ejecutable creado en bin/$(APP_NAME)"

clean:
	@echo "Limpiando..."
	@rm -rf bin/

run-linux: build-linux
	@echo "Ejecutando en Linux (requiere Wine o Linux)"
	@./bin/$(APP_NAME)