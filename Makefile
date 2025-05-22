# Makefile for Deno Telegram <-> Kafka (Redpanda) Bot

.PHONY: install setup run dev stop

# Redpanda config
KAFKA_BROKER=localhost:9092
KAFKA_IN_TOPIC?=com.sectorflabs.ratatoskr.in
KAFKA_OUT_TOPIC?=com.sectorflabs.ratatoskr.out

# Install Redpanda natively (macOS/Linux with Homebrew)
install:
	brew install redpanda-data/tap/redpanda

# Setup Redpanda natively and create topics
setup:
	rpk container start
	rpk topic create $(KAFKA_IN_TOPIC) || true; \
	rpk topic create $(KAFKA_OUT_TOPIC) || true; \
	echo "Redpanda and topics are ready."

run:
	deno run -A main.ts

dev:
	deno run --watch -A main.ts

stop:
	pkill -f "redpanda start" || true 

consume:
	rpk topic consume com.sectorflabs.ratatoskr.in --offset end -n 1 | jq

produce:
	echo '{"chat_id":$(CHAT_ID),"text":"hi from make"}' | rpk topic produce com.sectorflabs.ratatoskr.out

pushpi:
	rsync -av --delete \
		Cargo.toml Cargo.lock \
		src scripts \
		divanvisagie@heimdallr:~/src/ratatoskr
