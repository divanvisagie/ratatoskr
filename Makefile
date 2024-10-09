APP_NAME=ratatoskr

main:  
	go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

dev:
	go run cmd/$(APP_NAME)/main.go