GOOS=windows GOARCH=amd64 go build -o dist/windows main.go
GOOS=darwin GOARCH=amd64 go build -o dist/macos-intel main.go
GOOS=darwin GOARCH=arm64 go build -o dist/macos-m1 main.go
GOOS=linux GOARCH=amd64 go build -o dist/ubuntu main.go
