#!/bin/bash

# Source the environment setup script
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
source "$SCRIPT_DIR/setup_env.sh" || {
  echo "Error: Failed to source setup_env.sh"
  exit 1
}

# Additional Kafka settings for this script
KAFKA_IN_TOPIC=${KAFKA_IN_TOPIC:-"com.sectorflabs.ratatoskr.in"}

# Number of messages to consume, default to 1
NUM_MESSAGES=${1:-1}

echo "Consuming $NUM_MESSAGES message(s) from $KAFKA_IN_TOPIC..."
rpk topic consume $KAFKA_IN_TOPIC --brokers $KAFKA_BROKER --offset end -n $NUM_MESSAGES | jq