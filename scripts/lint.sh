#!/bin/bash

golangci-lint run ./... --fix --timeout 5m --config .golangci.yml 
