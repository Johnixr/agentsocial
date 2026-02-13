.PHONY: build dev web-install web-dev web-build all clean run migrate

# Go build settings
BINARY_NAME := agentsocial
BUILD_DIR := bin
CMD_DIR := ./cmd/server/

# Build the Go backend binary
build:
	@echo "Building backend..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the backend in development mode
dev:
	@echo "Starting backend in dev mode..."
	go run $(CMD_DIR)

# Install frontend dependencies
web-install:
	@echo "Installing frontend dependencies..."
	cd web && npm install

# Run the frontend dev server
web-dev:
	@echo "Starting frontend dev server..."
	cd web && npm run dev

# Build the frontend for production
web-build:
	@echo "Building frontend..."
	cd web && npm run build
	@echo "Frontend built: web/dist/"

# Build everything (backend + frontend)
all: web-install web-build build
	@echo "Full build complete."

# Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf web/dist
	rm -rf web/node_modules
	@echo "Clean complete."

# Run the built binary
run:
	@echo "Starting $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run database migration (auto-migrates on server startup)
migrate:
	@echo "Running migrations (starts server, auto-migrates on startup)..."
	go run $(CMD_DIR)
