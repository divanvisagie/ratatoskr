#!/bin/bash

# Ultra simple test script for Ratatoskr

# Load environment variables
source ~/.envrc 2>/dev/null || true

# Use environment CHAT_ID or default value if not set
CHAT_ID=${CHAT_ID:-123456789}
KAFKA_BROKER=${KAFKA_BROKER:-"localhost:9092"}
KAFKA_OUT_TOPIC=${KAFKA_OUT_TOPIC:-"com.sectorflabs.ratatoskr.out"}

# Prepare a simple message
MSG='{"chat_id":'$CHAT_ID',"text":"Test from ultra simple script"}'

# Send it directly to Kafka
echo "$MSG" | rpk topic produce $KAFKA_OUT_TOPIC --brokers $KAFKA_BROKER

# Print confirmation
echo "Simple message sent to $KAFKA_OUT_TOPIC"