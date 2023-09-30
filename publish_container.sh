#!/bin/bash

# Initialise buildx (Skip this if already done)
docker buildx create --use

# Login to GitHub Container Registry (Skip if already logged in)
echo $GH_TOKEN | docker login ghcr.io -u $GH_USERNAME --password-stdin

# Build and push multi-platform image
docker buildx build --platform linux/amd64,linux/arm64 \
  --push \
  -t ghcr.io/$GH_USERNAME/ratatoskr:latest .
