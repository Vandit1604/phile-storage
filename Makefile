# Binary name
BINARY_NAME=phile-storage

# Number of peers (default 3, override with `make run PEERS=5`)
PEERS?=3

# Peer ports (starting from 5001)
PORT_START=5001

# Docker services
ETCD_CONTAINER=etcd
REDIS_CONTAINER=redis

# Default target: Show help
.DEFAULT_GOAL := help

# Help command to display all available targets
help:
	@echo ""
	@echo "📌 Available Commands:"
	@echo "-----------------------------------------------"
	@echo "make build          - Build the Go binary"
	@echo "make run            - Start multiple peers (default: 3)"
	@echo "make run PEERS=N    - Start N peers"
	@echo "make stop           - Stop all running peers"
	@echo "make clean          - Remove build artifacts"
	@echo "make start-services - Start Etcd & Redis (Docker)"
	@echo "make stop-services  - Stop Etcd & Redis"
	@echo "make reset          - Stop everything & clean up"
	@echo "make help           - Show available commands"
	@echo "-----------------------------------------------"
	@echo ""

# Build the Go binary
build: ## Compile the Go project
	go build -o bin/$(BINARY_NAME) cmd/main.go

# Run multiple peers in background
run: ## Start multiple peers in background
	@echo "🚀 Starting $(PEERS) peers..."
	@for i in $$(seq 0 $$(($(PEERS)-1))); do \
		PORT=$$(($(PORT_START) + $$i)); \
		echo "Starting peer on port $$PORT..."; \
		PEER_IP="127.0.0.1" PEER_PORT="$$PORT" .bin/$(BINARY_NAME) & \
	done
	@echo "✅ All peers started!"

# Stop all running Go processes
stop: ## Stop all running peers
	@echo "🛑 Stopping all running peers..."
	@pkill -f $(BINARY_NAME) || true
	@echo "✅ All peers stopped!"

# Clean build artifacts
clean: ## Remove built binary files
	@echo "🧹 Cleaning build files..."
	@rm -f $(BINARY_NAME)
	@echo "✅ Cleaned!"

# Start required services (Etcd & Redis)
start-services: ## Start Etcd & Redis in Docker
	@echo "🚀 Starting Etcd & Redis..."
	@docker start $(ETCD_CONTAINER) || docker run -p 2379:2379 --name $(ETCD_CONTAINER) -d \
		quay.io/coreos/etcd:v3.5.19 etcd -advertise-client-urls http://0.0.0.0:2379 \
		-listen-client-urls http://0.0.0.0:2379
	@docker start $(REDIS_CONTAINER) || docker run --name $(REDIS_CONTAINER) -p 6379:6379 -d redis
	@echo "✅ Etcd & Redis running!"

# Stop Docker services
stop-services: ## Stop Etcd & Redis
	@echo "🛑 Stopping Etcd & Redis..."
	@docker stop $(ETCD_CONTAINER) $(REDIS_CONTAINER) || true
	@echo "✅ Services stopped!"

# Full reset (stop everything and clean)
reset: stop stop-services clean ## Stop everything & clean up
	@echo "🔄 Full reset completed!"

.PHONY: build run stop clean start-services stop-services reset help

