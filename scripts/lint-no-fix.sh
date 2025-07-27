#!/bin/bash

golangci-lint run ./... --timeout 5m --config .golangci.yml 
