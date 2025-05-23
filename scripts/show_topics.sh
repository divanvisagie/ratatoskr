#!/bin/bash

# Script to show contents of Ratatoskr Kafka topics
# Usage: ./show_topics.sh [number_of_messages]

# Source the environment setup script
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
source "$SCRIPT_DIR/setup_env.sh" || {
  echo "Error: Failed to source setup_env.sh"
  exit 1
}

# Number of messages to show per topic, default to 5
NUM_MESSAGES=${1:-5}

# Colors for terminal output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Show Kafka topic information
echo -e "${BLUE}=== Ratatoskr Kafka Topics Information ===${NC}"

# Show topics
echo -e "\n${GREEN}Available Kafka Topics:${NC}"
rpk topic list --brokers $KAFKA_BROKER

# Show IN topic contents
echo -e "\n${YELLOW}Contents of IN Topic ($KAFKA_IN_TOPIC):${NC}"
echo -e "Showing last $NUM_MESSAGES messages..."
rpk topic consume $KAFKA_IN_TOPIC --brokers $KAFKA_BROKER -n $NUM_MESSAGES -o end | jq '.' || echo "No messages or error reading topic"

# Show OUT topic contents
echo -e "\n${YELLOW}Contents of OUT Topic ($KAFKA_OUT_TOPIC):${NC}"
echo -e "Showing last $NUM_MESSAGES messages..."
rpk topic consume $KAFKA_OUT_TOPIC --brokers $KAFKA_BROKER -n $NUM_MESSAGES -o end | jq '.' || echo "No messages or error reading topic"

echo -e "\n${BLUE}=== Topic inspection complete ===${NC}"