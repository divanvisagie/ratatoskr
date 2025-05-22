#!/bin/bash

# Very simple producer script for Ratatoskr that handles quotes and special characters properly
# Usage: ./simple_produce.sh "Your message here"

# Source environment variables if .envrc exists
if [ -f .envrc ]; then
  source .envrc
fi

# Use environment CHAT_ID or show error if not set
if [ -z "$CHAT_ID" ]; then
  echo "Error: CHAT_ID environment variable is not set"
  echo "Please set it in .envrc or export it before running this script"
  exit 1
fi

# Set Kafka configuration
KAFKA_BROKER=${KAFKA_BROKER:-"localhost:9092"}
KAFKA_OUT_TOPIC=${KAFKA_OUT_TOPIC:-"com.sectorflabs.ratatoskr.out"}

# Get message text from argument or use default
MESSAGE_TEXT=${1:-"Hello from simple_produce script!"}

# Create a temporary file for the message
TMP_FILE=$(mktemp)

# Write the correctly formatted JSON to the temporary file
cat > "$TMP_FILE" << EOF
{"chat_id":$CHAT_ID,"text":"$MESSAGE_TEXT"}
EOF

# Display the message being sent
echo "Sending message to $KAFKA_OUT_TOPIC:"
cat "$TMP_FILE"

# Send the message to Kafka directly from the file
cat "$TMP_FILE" | rpk topic produce "$KAFKA_OUT_TOPIC" --brokers "$KAFKA_BROKER"

# Clean up
rm "$TMP_FILE"

echo "Message sent!"