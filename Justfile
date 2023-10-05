#!/usr/bin/env just

test:
	go test -timeout 3s ./...
