.PHONEY: build
build:
	GOOS=windows GOARCH=amd64 go build -o cmd/kubeslice-windows-amd64.exe main.go
	GOOS=linux GOARCH=amd64 go build -o cmd/kubeslice-linux-amd64 main.go
	GOOS=linux GOARCH=arm go build -o cmd/kubeslice-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o cmd/kubeslice-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o cmd/kubeslice-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o cmd/kubeslice-darwin-arm64 main.go