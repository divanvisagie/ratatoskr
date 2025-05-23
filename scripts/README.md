# Ratatoskr Test Scripts

This directory contains scripts for testing and interacting with Ratatoskr's Kafka message handling. These scripts are designed to help developers test message handling without needing to set up a full Telegram environment.

## Prerequisites

- [Redpanda CLI (`rpk`)](https://docs.redpanda.com/docs/reference/rpk-commands/) or Kafka CLI tools
- `jq` for JSON formatting (install with `brew install jq` on macOS or `apt install jq` on Debian/Ubuntu)

## Available Scripts

### 1. Show Topic Contents

```bash
./scripts/show_topics.sh [number_of_messages]
```

Displays information about both Kafka topics, including their current contents.

**Arguments:**
- `number_of_messages`: Optional. Number of messages to show from each topic (default: 5)

### 2. Consume Messages

```bash
./scripts/consume.sh [number_of_messages]
```

Consumes messages from the input Kafka topic (`com.sectorflabs.ratatoskr.in` by default).

**Arguments:**
- `number_of_messages`: Optional. Number of messages to consume (default: 1)

### 3. Produce Basic Messages

```bash
./scripts/produce.sh [message_text]
```

Produces a simple text message to the output Kafka topic (`com.sectorflabs.ratatoskr.out` by default).

**Arguments:**
- `message_text`: Optional. The text content of the message (default: "Hello from Ratatoskr test script!")

**Note:** This script uses the `CHAT_ID` environment variable from your `.envrc` file.

### 4. Produce Messages with Buttons

```bash
./scripts/produce_with_buttons.sh [message_text]
```

Produces a message with inline buttons to the output Kafka topic.

**Arguments:**
- `message_text`: Optional. The text content of the message (default: "This is a test message with buttons!")

**Note:** This script uses the `CHAT_ID` environment variable from your `.envrc` file.

### 5. Simulate Callback Query (Button Click)

```bash
./scripts/simulate_callback.sh [message_id] [callback_data] [callback_query_id]
```

Simulates a button click by producing a callback query message to the input Kafka topic.

**Arguments:**
- `message_id`: Optional. The message ID containing the clicked button (default: "1234")
- `callback_data`: Optional. The callback data associated with the clicked button (default: "button1_action")
- `callback_query_id`: Optional. A unique ID for this callback (generated automatically if not provided)

**Note:** This script uses the `CHAT_ID` environment variable from your `.envrc` file for both the chat_id and user_id (unless USER_ID is separately defined).

### 6. Test Full Flow

```bash
./scripts/test_full_flow.sh
```

Runs a complete test flow that demonstrates the entire message cycle:
1. Sending a message with buttons
2. Simulating receipt of the message in Telegram
3. Simulating a button click
4. Consuming the callback message

**Note:** This script uses the `CHAT_ID` environment variable from your `.envrc` file.

## Environment Variables

All scripts respect the following environment variables:

- `CHAT_ID`: Telegram chat ID to use (required, normally defined in `.envrc`)
- `USER_ID`: Telegram user ID (defaults to same value as CHAT_ID if not provided)
- `KAFKA_BROKER`: Kafka broker address (default: "localhost:9092")
- `KAFKA_IN_TOPIC`: Input topic name (default: "com.sectorflabs.ratatoskr.in")
- `KAFKA_OUT_TOPIC`: Output topic name (default: "com.sectorflabs.ratatoskr.out")

You can override these by setting the variables before running the scripts:

```bash
KAFKA_BROKER=my-kafka:9093 KAFKA_OUT_TOPIC=custom.topic ./scripts/produce.sh "Custom message"
```

## Making Scripts Executable

Before running the scripts, make them executable:

```bash
chmod +x scripts/*.sh
```