APP_NAME=ratatoskr
BUILD_DIR=bin

main:
	go build cmd/ratatoskr/main.go

pushpi:
	ssh heimdallr.local "mkdir -p ~/src/" && rsync -av --progress . heimdallr.local:~/src/$(APP_NAME)

run:
	go run cmd/ratatoskr/main.go

staging.run:
	docker compose -f docker-compose.staging.yml up

prod.run:
	docker compose -f docker-compose.prod.yml up

clean:
	rm -rf ratatoskr

test:
	go test -v ./...

