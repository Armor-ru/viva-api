run:
	go mod download
	go run cmd/viva-api/main.go -cf ./config/viva-api.yaml

build:
	GOOS=linux
	go mod download
	go build -o ./.dist/viva-api ./cmd/viva-api/main.go
