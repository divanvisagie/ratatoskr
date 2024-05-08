APP_NAME=ratatoskr

main:
	cargo build --release

# run docker compose on docker-compose.test.yml
test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

pushpi:
	ssh $(PI) "mkdir -p ~/src/" \
	&& rsync -av --progress src $(PI):~/src/$(APP_NAME) \
    && rsync -av --progress Cargo.toml $(PI):~/src/$(APP_NAME) \
	&& rsync -av --progress Cargo.lock $(PI):~/src/$(APP_NAME) \
	&& rsync -av --progress Makefile $(PI):~/src/$(APP_NAME) \

install:
	# stop the service if it already exists
	systemctl stop ratatoskr || true
	systemctl disable ratatoskr || true
	# delete the old service file if it exists
	rm /etc/systemd/system/ratatoskr.service || true
	cp scripts/ratatoskr.service /etc/systemd/system/

dev:
	cargo watch -x run
