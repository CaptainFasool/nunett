#!/bin/bash

go test cli_test.go
go test -ldflags="-extldflags=-Wl,-z,lazy" -run=TestSecurity
