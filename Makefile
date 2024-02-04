BIN_NAME = ntpc

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./$(BIN_NAME) -tags linux .

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./$(BIN_NAME) -tags linux .

build-macos-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./$(BIN_NAME) -tags darwin .

build-macos-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./$(BIN_NAME) -tags darwin .
