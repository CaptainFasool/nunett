`./test` directory of the package contains full test suite of DMS. 

## Run CLI Test Suite

This command will run the Command Line Interface Test suite inside the `./test` directory:
`go test -run `

## Run Security Test Suite

This command will run the Security Test suite inside the `./test` directory:
`go test -ldflags="-extldflags=-Wl,-z,lazy" -run=TestSecurity`

## Run all tests

This command will run all tests from root directory:
`sh run_all.sh`

After developing a new test suite or a test make sure that they are properly included with approprate flags and parameters into the `run_all.sh` file.

