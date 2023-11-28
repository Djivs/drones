#!/bin/sh

export PATH=$(go env GOPATH)/bin:$PATH

swag init -g cmd/drones/main.go