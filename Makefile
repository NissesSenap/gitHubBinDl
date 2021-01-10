all: lint fmt vet build

lint:
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

build:
	GOOS=linux GOARCH=amd64 go build -ldflags '-extldflags "-static"' -o bin/githubbindl_linux_amd64 cmd/githubbindl/main.go

build/windows:
	GOOS=windows GOARCH=amd64 go build -o bin/githubbindl_windows_amd64.exe cmd/githubbindl/main.go

build/mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/githubbindl_darwin_amd64 cmd/githubbindl/main.go