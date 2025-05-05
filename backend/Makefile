# Variables
BINARY_NAME = phile-storage
DOCKER_COMPOSE_FILE = docker-compose.yml
PEER_COUNT ?= 1

# Help message
help:
	@echo "Available commands:"
	@echo "  make build        - Build the Go binary"
	@echo "  make run          - Run the application with default PEER_COUNT=1"
	@echo "  make run-peers N  - Run the application with N peers (e.g., make run-peers PEER_COUNT=3)"
	@echo "  make start-docker - Start necessary Docker containers (Etcd & Redis)"
	@echo "  make stop-docker  - Stop Docker containers"
	@echo "  make stop-all     - Stop everything (Docker + app)"
	@echo "  make clean        - Remove built binary"

# Start required Docker services (Etcd & Redis)
start-docker:
	@echo "🚀 Starting Etcd & Redis containers..."
	docker start etcd || docker run -d --name etcd -p 2379:2379 quay.io/coreos/etcd:v3.5.19 etcd -advertise-client-urls http://0.0.0.0:2379 -listen-client-urls http://0.0.0.0:2379
	docker start redis || docker run -d --name redis -p 6379:6379 redis

# Stop Docker services
stop-docker:
	@echo "🛑 Stopping Etcd & Redis containers..."
	docker stop etcd redis || true

# Build the Go binary
build:
	@echo "🏗️  Building the Go binary..."
	go build -o ./bin/$(BINARY_NAME) cmd/main.go

# Run the application with default PEER_COUNT=1
run: build
	@echo "🚀 Running application with $(PEER_COUNT) peers..."
	./bin/$(BINARY_NAME) -peers=$(PEER_COUNT)

# Run with a specific number of peers (e.g., make run-peers PEER_COUNT=3)
run-peers: build
	@echo "🚀 Running application with $(PEER_COUNT) peers..."
	./bin/$(BINARY_NAME) -peers=$(PEER_COUNT)

# Stop everything (Docker & app)
stop-all: stop-docker
	@echo "🛑 Stopping all running application processes..."
	pkill -f $(BINARY_NAME) || true

# Remove built binary
clean:
	@echo "🗑️  Cleaning up..."
	rm -f $(BINARY_NAME)

