VERSION := 0.1.0

# Default target
.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "🚀 Sylfie Compiler - Available Make Targets"
	@echo ""
	@echo "🔨 Development:"
	@echo "  run              - Compile and run examples recursively"
	@echo "  watch            - Watch files for changes and recompile"
	@echo "  build            - Build the sylfie binary"
	@echo "  fmt              - Format Go code"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test                        - Run all Go unit tests"
	@echo "  test-codegen                - Run code generation tests"
	@echo "  test-view-transformer       - Run view transformer tests"
	@echo "  test-view-transformer-update - Update view transformer golden files"
	@echo "  test-all                    - Run all tests (unit + golden file tests)"
	@echo ""
	@echo "✨ Golden File Tests (End-to-End):"
	@echo "  test-golden                     - Run all golden file tests"
	@echo "  test-golden-update              - Update all golden files"
	@echo "  test-golden-list                - Show available test categories"
	@echo "  test-golden-category            - Run tests for specific category (use CATEGORY=name)"
	@echo "  test-golden-category-update     - Update golden files for category (use CATEGORY=name)"
	@echo "  test-golden-single              - Run single test (use TEST=category/testname)"
	@echo "  test-golden-single-update       - Update single golden file (use TEST=category/testname)"
	@echo "  test-golden-diff                - Compare generated vs expected files (use CATEGORY=name)"
	@echo "  test-golden-diff-single         - Compare single generated vs expected file (use TEST=name)"
	@echo "  test-golden-clean               - Clean generated test files"
	@echo ""
	@echo "🔍 Compiler Operations:"
	@echo "  scan             - Scan source files and show tokens"
	@echo "  parse            - Parse source files and show AST"
	@echo "  compile          - Compile a single source file"
	@echo ""
	@echo "🌐 Web Development:"
	@echo "  web              - Start development web server"
	@echo ""
	@echo "📝 Quick Examples:"
	@echo "  make test-golden-list                           # See all test categories"
	@echo "  make test-golden-category CATEGORY=basic       # Test basic functionality"
	@echo "  make test-golden-single TEST=views/simple_view # Test one specific file"

.PHONY: run
run:
	go run ./cmd/... compile examples/ -r -d

.PHONY: test
test:
	go test ./...

.PHONY: test-codegen
test-codegen:
	@echo "🧪 Running codegen tests..."
	@go test ./compiler/codegen -v

.PHONY: test-view-transformer
test-view-transformer:
	@echo "🧪 Running view transformer tests..."
	@go test ./compiler/transformers -v

.PHONY: test-view-transformer-update
test-view-transformer-update:
	@echo "📝 Updating view transformer golden files..."
	@UPDATE_GOLDEN=1 go test ./compiler/transformers -v

.PHONY: watch
watch:
	go run ./cmd/... watch examples/sylfie/01_hello_world/ -d

.PHONY: scan
scan:
	go run ./cmd/... scan examples/sylfie/ -d -w -r

.PHONY: parse
parse:
	go run ./cmd/... parse examples/sylfie/ -d -w -r

.PHONY: compile
compile:
	go run ./cmd/... compile examples/sylfie/01_hello_world/01_hello_world.psx -d -r

.PHONY: web
web:
	poetry run uvicorn examples.sylfie.01_hello_world.01_hello_world:app --reload

build:
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/sylfie ./cmd/...

.PHONY: fmt	
fmt:
	go fmt ./...

# Golden file tests (end-to-end compiler tests)
.PHONY: test-golden
test-golden:
	@echo "🧪 Running golden file tests..."
	@go test ./compiler -run "TestE2E" -v

.PHONY: test-golden-update
test-golden-update:
	@echo "📝 Updating golden files..."
	@UPDATE_GOLDEN=1 go test ./compiler -run "TestE2E" -v
	@echo "✅ Golden files updated successfully!"

.PHONY: test-golden-category
test-golden-category:
	@if [ -z "$(CATEGORY)" ]; then \
		echo "❌ Error: Please specify a category with CATEGORY=<name>"; \
		echo "Available categories: basic, views, control_flow, composition, slots, attributes, expressions, python_integration, htmx, fastapi, errors"; \
		exit 1; \
	fi
	@echo "🧪 Running golden file tests for category: $(CATEGORY)"
	@go test ./compiler -run "TestE2E/$(CATEGORY)" -v

.PHONY: test-golden-category-update
test-golden-category-update:
	@if [ -z "$(CATEGORY)" ]; then \
		echo "❌ Error: Please specify a category with CATEGORY=<name>"; \
		echo "Available categories: basic, views, control_flow, composition, slots, attributes, expressions, python_integration, htmx, fastapi, errors"; \
		exit 1; \
	fi
	@echo "📝 Updating golden files for category: $(CATEGORY)"
	@UPDATE_GOLDEN=1 go test ./compiler -run "TestE2E/$(CATEGORY)" -v
	@echo "✅ Golden files updated for $(CATEGORY)!"

.PHONY: test-golden-single
test-golden-single:
	@if [ -z "$(TEST)" ]; then \
		echo "❌ Error: Please specify a test with TEST=<category>/<testname>"; \
		echo "Example: make test-golden-single TEST=basic/hello_world"; \
		exit 1; \
	fi
	@echo "🧪 Running single golden file test: $(TEST)"
	@go test ./compiler -run "TestE2E/$(TEST)" -v

.PHONY: test-golden-single-update
test-golden-single-update:
	@if [ -z "$(TEST)" ]; then \
		echo "❌ Error: Please specify a test with TEST=<category>/<testname>"; \
		echo "Example: make test-golden-single-update TEST=basic/hello_world"; \
		exit 1; \
	fi
	@echo "📝 Updating single golden file: $(TEST)"
	@UPDATE_GOLDEN=1 go test ./compiler -run "TestE2E/$(TEST)" -v
	@echo "✅ Golden file updated for $(TEST)!"

.PHONY: test-golden-list
test-golden-list:
	@echo "📋 Available golden file test categories:"
	@echo "  • basic - Simple Python code and basic views"
	@echo "  • views - View definitions and parameters"
	@echo "  • control_flow - If statements, loops, match, try/except"
	@echo "  • composition - View composition and nesting"
	@echo "  • slots - Slot functionality and templates"
	@echo "  • attributes - HTML attribute handling"
	@echo "  • expressions - Complex expressions and f-strings"
	@echo "  • python_integration - Python-specific features"
	@echo "  • htmx - HTMX integration examples"
	@echo "  • fastapi - FastAPI integration examples"
	@echo "  • errors - Error handling and edge cases"
	@echo ""
	@echo "📝 Usage examples:"
	@echo "  make test-golden                                    # Run all golden file tests"
	@echo "  make test-golden-update                             # Update all golden files"
	@echo "  make test-golden-category CATEGORY=basic           # Test specific category"
	@echo "  make test-golden-category-update CATEGORY=views    # Update specific category"
	@echo "  make test-golden-single TEST=basic/hello_world     # Test single file"
	@echo "  make test-golden-single-update TEST=views/simple   # Update single file"
	@echo ""
	@echo "🔍 File comparison (generated vs expected):"
	@echo "  make test-golden-diff CATEGORY=basic               # Compare all files in category"
	@echo "  make test-golden-diff-single TEST=basic/hello      # Compare single file"

.PHONY: test-golden-diff
test-golden-diff:
	@if [ -z "$(CATEGORY)" ]; then \
		echo "❌ Error: Please specify a category with CATEGORY=<name>"; \
		echo "Available categories: basic, views, control_flow, composition, slots, attributes, expressions, python_integration, htmx, fastapi, errors"; \
		exit 1; \
	fi
	@echo "🔍 Comparing generated vs expected files for category: $(CATEGORY)"
	@if [ ! -d "compiler/testdata/generated/$(CATEGORY)" ]; then \
		echo "⚠️  Generated directory not found. Run tests first: make test-golden-category CATEGORY=$(CATEGORY)"; \
		exit 1; \
	fi
	@if command -v diff >/dev/null 2>&1; then \
		echo "Running diff between generated and expected files:"; \
		diff -u "compiler/testdata/expected/$(CATEGORY)" "compiler/testdata/generated/$(CATEGORY)" || true; \
	else \
		echo "Diff command not available. Files are in:"; \
		echo "  Expected: compiler/testdata/expected/$(CATEGORY)"; \
		echo "  Generated: compiler/testdata/generated/$(CATEGORY)"; \
	fi

.PHONY: test-golden-diff-single
test-golden-diff-single:
	@if [ -z "$(TEST)" ]; then \
		echo "❌ Error: Please specify a test with TEST=<category>/<testname>"; \
		echo "Example: make test-golden-diff-single TEST=basic/hello_world"; \
		exit 1; \
	fi
	@CATEGORY=$$(echo "$(TEST)" | cut -d'/' -f1); \
	TESTNAME=$$(echo "$(TEST)" | cut -d'/' -f2); \
	EXPECTED_FILE="compiler/testdata/expected/$$CATEGORY/$$TESTNAME.py"; \
	GENERATED_FILE="compiler/testdata/generated/$$CATEGORY/$$TESTNAME.py"; \
	echo "🔍 Comparing generated vs expected for: $(TEST)"; \
	if [ ! -f "$$GENERATED_FILE" ]; then \
		echo "⚠️  Generated file not found. Run test first: make test-golden-single TEST=$(TEST)"; \
		exit 1; \
	fi; \
	if [ ! -f "$$EXPECTED_FILE" ]; then \
		echo "⚠️  Expected file not found: $$EXPECTED_FILE"; \
		exit 1; \
	fi; \
	if command -v diff >/dev/null 2>&1; then \
		echo "Expected file: $$EXPECTED_FILE"; \
		echo "Generated file: $$GENERATED_FILE"; \
		echo "Differences:"; \
		diff -u "$$EXPECTED_FILE" "$$GENERATED_FILE" || true; \
	else \
		echo "Diff command not available. Files are:"; \
		echo "  Expected: $$EXPECTED_FILE"; \
		echo "  Generated: $$GENERATED_FILE"; \
	fi

.PHONY: test-golden-clean
test-golden-clean:
	@echo "🧹 Cleaning generated test files..."
	@rm -rf compiler/testdata/generated/
	@echo "✅ Generated files cleaned!"

.PHONY: test-all
test-all: test test-golden
	@echo "✅ All tests completed successfully!"

# Legacy aliases for backwards compatibility
.PHONY: test-e2e test-e2e-update
test-e2e: test-golden
test-e2e-update: test-golden-update