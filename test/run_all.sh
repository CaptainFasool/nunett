#!/bin/bash

go test  -run=TestCLI
go test -ldflags="-extldflags=-Wl,-z,lazy" -run=TestSecurity`

