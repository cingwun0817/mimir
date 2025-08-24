LDFLAGS = -s -w

build:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/app_amd64 main.go
	go build -ldflags "$(LDFLAGS)" -o bin/app_mac main.go