main:
	go build cmd/ratatoskr/main.go


run:
	go run cmd/ratatoskr/main.go


test:
	go test -v ./...