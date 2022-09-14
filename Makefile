.PHONEY: build
build:
	GOOS=windows GOARCH=amd64 go build -o bin/kubeslice-cli-windows-amd64.exe main.go
	GOOS=linux GOARCH=amd64 go build -o bin/kubeslice-cli-linux-amd64 main.go
	GOOS=linux GOARCH=arm go build -o bin/kubeslice-cli-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/kubeslice-cli-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/kubeslice-cli-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/kubeslice-cli-darwin-arm64 main.go