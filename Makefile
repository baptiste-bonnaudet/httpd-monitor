.PHONY: help

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}';

build: ## Build the test environment 
	docker-compose build

up: ## Start the test environment 
	docker-compose up -d 

down: ## Start the test environment 
	docker-compose down --remove-orphans; 

logs: ## Show and follow the containers logs
	docker-compose logs -f;
