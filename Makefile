build:
	go build -o bin/main main.go
linux:
	GOOS=linux GOARCH=amd64 go build -o  bin/linux main.go
