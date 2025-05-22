#!/bin/bash

# Test Full Flow Script for Ratatoskr
# This script demonstrates a complete test cycle of:
# 1. Producing a message with buttons
# 2. Consuming the output from Ratatoskr
# 3. Simulating a button click (callback)
# 4. Consuming the callback in Kafka

# Source the environment setup script
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
source "$SCRIPT_DIR/setup_env.sh" || {
  echo "Error: Failed to source setup_env.sh"
  exit 1
}

# Set default values and color codes
# Use environment variables if they exist, otherwise use defaults
USER_ID=${USER_ID:-"$CHAT_ID"}  # Default to using the same ID as the chat
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Introduction
echo -e "${BLUE}=== Ratatoskr Full Test Flow ===${NC}"
echo -e "This script will test the full message flow using chat_id: ${YELLOW}$CHAT_ID${NC} and user_id: ${YELLOW}$USER_ID${NC}"
echo ""

# Step 1: Produce a message with buttons
echo -e "${GREEN}STEP 1: Sending a message with buttons${NC}"
echo -e "Executing: ./scripts/produce_with_buttons.sh \"Testing the full message flow\""
./scripts/produce_with_buttons.sh "Testing the full message flow"
echo ""
sleep 2

# Step 2: Wait for user to see the message in Telegram
echo -e "${YELLOW}The message should now be visible in Telegram.${NC}"
echo -e "Press Enter after you've confirmed the message was received..."
read

# Step 3: Simulate a button click
echo -e "${GREEN}STEP 3: Simulating a button click${NC}"
MESSAGE_ID="1234" # In a real scenario, you'd need to know the actual message ID
CALLBACK_DATA="button1_action"
echo -e "Executing: ./scripts/simulate_callback.sh $MESSAGE_ID $CALLBACK_DATA"
./scripts/simulate_callback.sh $MESSAGE_ID $CALLBACK_DATA
echo ""
sleep 2

# Step 4: Consume messages from the input topic
echo -e "${GREEN}STEP 4: Consuming callback message from Kafka${NC}"
echo -e "Executing: ./scripts/consume.sh 1"
./scripts/consume.sh 1
echo ""

# Conclusion
echo -e "${BLUE}=== Test Flow Complete ===${NC}"
echo "This test demonstrates the bidirectional flow:"
echo "1. A message with buttons sent to Telegram"
echo "2. A simulated button click from Telegram"
echo "Both directions are working if all steps completed successfully."
echo ""
echo -e "${YELLOW}Note: In a real application, Ratatoskr would process these messages.${NC}"