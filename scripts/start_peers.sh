#!/bin/bash

for port in {5001..5005}; do
  PEER_IP="127.0.0.1" PEER_PORT="$port" go run cmd/main.go &
done
