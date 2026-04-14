.PHONY: help backend frontend test clean dev

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

backend: ## Build Go backend
	cd checkers-solver/backend && go build -o ../../bin/solver ./...

frontend: ## Build frontend
	cd checkers-solver/frontend && npm run build

test: ## Run all tests
	cd checkers-solver/backend && go test ./...

dev: ## Start frontend dev server (proxies to backend)
	cd checkers-solver/frontend && npm run dev

clean: ## Remove build artifacts
	rm -rf bin/
	rm -rf checkers-solver/frontend/dist/
