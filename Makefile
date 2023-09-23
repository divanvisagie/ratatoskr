APP_NAME=ratatoskr
BUILD_DIR=bin

main:
	go build cmd/ratatoskr/main.go

build.arm:
	ssh heimdallr.local "mkdir -p ~/src/" && scp -r . heimdallr.local:~/src/$(APP_NAME)
	ssh heimdallr.local "cd ~/src/ratatoskr && make"

run:
	go run cmd/ratatoskr/main.go

clean:
	rm -rf ratatoskr

test:
	go test -v ./...
