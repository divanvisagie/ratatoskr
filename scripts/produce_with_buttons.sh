#!/bin/bash

# Source the environment setup script
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
source "$SCRIPT_DIR/setup_env.sh" || {
  echo "Error: Failed to source setup_env.sh"
  exit 1
}

# Additional Kafka settings for this script
KAFKA_OUT_TOPIC=${KAFKA_OUT_TOPIC:-"com.sectorflabs.ratatoskr.out"}

# Set a default message text
MESSAGE_TEXT=${1:-"This is a test message with buttons!"}

# Create a temporary file for the message
TMP_FILE=$(mktemp)

# Write the correctly formatted JSON to the temporary file
cat > "$TMP_FILE" << EOF
{"chat_id":$CHAT_ID,"text":"$MESSAGE_TEXT","buttons":[[{"text":"Button 1","callback_data":"button1_action"},{"text":"Button 2","callback_data":"button2_action"}],[{"text":"Button 3","callback_data":"button3_action"}]]}
EOF

# Display the message being sent
echo "Producing message with buttons to $KAFKA_OUT_TOPIC:"
cat "$TMP_FILE" | jq 2>/dev/null || cat "$TMP_FILE"

# Send the message to Kafka directly from the file
cat "$TMP_FILE" | rpk topic produce "$KAFKA_OUT_TOPIC" --brokers "$KAFKA_BROKER"

# Clean up
rm "$TMP_FILE"

echo "Message with buttons sent!"
