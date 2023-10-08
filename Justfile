#!/usr/bin/env just

all: format test lint

test:
  @echo Running Tests...
  @go test -timeout 3s ./...

format:
  @echo Formatting Files...
  @for file in $(find . -name "*.go"); do \
  gofumpt -w $file; \
  done

lint:
  @echo Running Linter...
  @golangci-lint run --max-issues-per-linter=0 --max-same-issues=0 ./...
