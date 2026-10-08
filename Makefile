.PHONY: init


GREEN=\033[0;32m
YELLOW=\033[0;33m
CYAN=\033[0;36m
NC=\033[0m


init:
	@echo "\$(CYAN)==== Setup local environment ====\$(NC)"
	
	
	@chmod +x .git-config/hooks/pre-commit
	@chmod +x .git-config/hooks/commit-msg
	
	@git config --local include.path ../.git-config/aliases
	@echo "\$(GREEN)+ aliases are set up succesfuly\$(NC)"
	
	@git config --local core.hooksPath ./.git-config/hooks
	@echo "\$(GREEN)+ hooks are set up succesfuly\$(NC)"
	
	@echo "\$(YELLOW)Ready for development, test & use. Try 'docker compose up --build'\$(NC)"

todo:	
	@echo "===-==-==-=- TODO -=-==-==-==="
	@cat .TODO
generate:
	@echo "\$(GREEN)Starting generate proto/account/v1/account.pb.go...\$(NC)"
	@protoc --plugin=protoc-gen-go=$(shell go env GOPATH)/bin/protoc-gen-go \
	       --plugin=protoc-gen-go-grpc=$(shell go env GOPATH)/bin/protoc-gen-go-grpc \
	       --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/account/v1/account.proto
	@echo "Account Service - incapsulated service without HTTP."
	@echo "All him API works by gRPC. Account Service link other services to Postgres"
	@echo "\$(GREEN)Finish without errors!\$(NC)"

