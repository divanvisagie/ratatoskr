APP_NAME=ratatoskr

main:
	cargo build --release

# run docker compose on docker-compose.test.yml
test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

pushpi:
	ssh heimdallr.local "mkdir -p ~/src/" && rsync -av --progress . heimdallr.local:~/src/$(APP_NAME)

install:
	# stop the service if it already exists
	systemctl stop ratatoskr || true
	systemctl disable ratatoskr || true
	# delete the old service file if it exists
	rm /etc/systemd/system/ratatoskr.service || true
	cp scripts/ratatoskr.service /etc/systemd/system/
