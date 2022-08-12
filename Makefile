.PHONEY: build
build:
	GOOS=windows GOARCH=amd64 go build -o bin/slicectl-windows-amd64.exe main.go
	GOOS=linux GOARCH=amd64 go build -o bin/slicectl-linux-amd64 main.go
	GOOS=linux GOARCH=arm go build -o bin/slicectl-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/slicectl-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/slicectl-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/slicectl-darwin-arm64 main.go