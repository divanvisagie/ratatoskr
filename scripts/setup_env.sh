#!/bin/bash

# Ensure environment is properly set up for Ratatoskr scripts
# This script can be sourced by other scripts to verify environment

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check for required tools
if ! command -v rpk &> /dev/null; then
    echo -e "${RED}Error: rpk (Redpanda CLI) not found${NC}"
    echo "Please install it with: brew install redpanda-data/tap/redpanda"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq not found${NC}"
    echo "Please install it with: brew install jq (macOS) or apt install jq (Debian/Ubuntu)"
    exit 1
fi

# Set default values for variables
KAFKA_BROKER=${KAFKA_BROKER:-"localhost:9092"}
KAFKA_IN_TOPIC=${KAFKA_IN_TOPIC:-"com.sectorflabs.ratatoskr.in"}
KAFKA_OUT_TOPIC=${KAFKA_OUT_TOPIC:-"com.sectorflabs.ratatoskr.out"}

# Check if CHAT_ID is set
if [ -z "$CHAT_ID" ]; then
    echo -e "${YELLOW}Warning: CHAT_ID environment variable is not set${NC}"
    echo "Please set it to your Telegram chat ID, for example:"
    echo "export CHAT_ID=123456789"
    echo "This is needed for most test scripts to work properly."
    
    # Allow script to continue if this was just sourced (not directly executed)
    if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
        exit 1
    fi
fi

# Print environment settings if being run directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo -e "${BLUE}=== Ratatoskr Environment ===${NC}"
    echo -e "CHAT_ID: ${YELLOW}$CHAT_ID${NC}"
    echo -e "KAFKA_BROKER: ${GREEN}$KAFKA_BROKER${NC}"
    echo -e "KAFKA_IN_TOPIC: ${GREEN}$KAFKA_IN_TOPIC${NC}"
    echo -e "KAFKA_OUT_TOPIC: ${GREEN}$KAFKA_OUT_TOPIC${NC}"
    
    # Check if topics exist
    echo -e "\n${BLUE}Available Kafka Topics:${NC}"
    rpk topic list --brokers $KAFKA_BROKER || {
        echo -e "${RED}Failed to list Kafka topics${NC}"
        echo "Make sure Kafka is running at $KAFKA_BROKER"
        exit 1
    }
    
    echo -e "\n${GREEN}Environment is ready for Ratatoskr scripts${NC}"
fi