# Makefile for Deno Telegram <-> Kafka (Redpanda) Bot

.PHONY: install setup run dev stop consume produce test_flow test_buttons test_callback show_topics

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
	./scripts/consume.sh $(N)

produce:
	./scripts/produce.sh "$(TEXT)"

test_buttons:
	./scripts/produce_with_buttons.sh "$(TEXT)"

test_callback:
	./scripts/simulate_callback.sh $(MESSAGE_ID) "$(CALLBACK_DATA)"

test_flow:
	./scripts/test_full_flow.sh

show_topics:
	./scripts/show_topics.sh $(N)

pushpi:
	rsync -av --delete \
		Cargo.toml Cargo.lock \
		src scripts \
		divanvisagie@heimdallr:~/src/ratatoskr
