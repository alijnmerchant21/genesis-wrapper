#!/usr/bin/make -f

install: go.sum
	@echo "Installing wrapper..."
	go install ./cmd/wrapper