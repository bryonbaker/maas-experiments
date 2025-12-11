.PHONY: clean build

# Binary name
BINARY_NAME=tier-admin

# Build directory
BUILD_DIR=.

# Source files
MAIN_SRC=cmd/server/main.go

clean:
	@echo "Cleaning..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Clean complete"

build: swagger
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_SRC)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag > /dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs; \
	elif [ -f $(HOME)/go/bin/swag ]; then \
		$(HOME)/go/bin/swag init -g cmd/server/main.go -o docs; \
	else \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		$(HOME)/go/bin/swag init -g cmd/server/main.go -o docs; \
	fi
	@echo "Swagger documentation updated"

