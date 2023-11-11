.PHONY: all clean linux_amd64 darwin_arm64

all: linux_amd64 darwin_arm64

linux_amd64:
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 go build -o builds/dms_linux_amd64 .

darwin_arm64:
	@echo "Building for Darwin ARM64..."
	GOOS=darwin GOARCH=arm64 go build -o builds/dms_darwin_arm64 .

darwin_amd64:
	@echo "Building for Darwin AMD64..."
	GOOS=darwin GOARCH=amd64 go build -o builds/dms_darwin_amd64 .

clean:
	@echo "Cleaning up..."
	rm -rf builds/
