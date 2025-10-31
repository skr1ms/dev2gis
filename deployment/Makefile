.PHONY: help deploy-prod deploy-dev rollback-prod rollback-dev logs ps restart scale up down

SERVICES=postgres rabbitmq minio nodeodm go-orchestrator python-heightmap frontend nginx prometheus grafana loki

help:
	@echo "Available targets:"
	@echo "  deploy-prod           - Deploy to production (requires COMMIT=<sha>)"
	@echo "  deploy-dev            - Deploy to development (requires COMMIT=<sha>)"
	@echo "  rollback-prod         - Rollback production deployment"
	@echo "  rollback-dev          - Rollback development deployment"
	@echo "  logs                  - Show logs for service (requires svc=<service>)"
	@echo "  ps                    - Show running containers"
	@echo "  restart               - Restart service (requires svc=<service>)"
	@echo "  scale                 - Scale service (requires svc=<service> n=<count>)"
	@echo "  up                    - Start all services"
	@echo "  down                  - Stop all services"
	@echo ""
	@echo "Services: $(SERVICES)"

deploy-prod:
	@[ -n "$(COMMIT)" ] || (echo "Usage: make deploy-prod COMMIT=<sha>"; exit 1)
	bash deployment/scripts/deploy.sh --env prod --commit $(COMMIT)

deploy-dev:
	@[ -n "$(COMMIT)" ] || (echo "Usage: make deploy-dev COMMIT=<sha>"; exit 1)
	bash deployment/scripts/deploy.sh --env dev --commit $(COMMIT)

rollback-prod:
	bash deployment/scripts/rollback.sh --env prod

rollback-dev:
	bash deployment/scripts/rollback.sh --env dev

logs:
	@[ -n "$(svc)" ] || (echo "Usage: make logs svc=<service>"; exit 1)
	docker compose -f deployment/docker/docker-compose.yml logs -f $(svc)

ps:
	docker compose -f deployment/docker/docker-compose.yml ps

restart:
	@[ -n "$(svc)" ] || (echo "Usage: make restart svc=<service>"; exit 1)
	docker compose -f deployment/docker/docker-compose.yml restart $(svc)

scale:
	@[ -n "$(svc)" ] || (echo "Usage: make scale svc=<service> n=<count>"; exit 1)
	@[ -n "$(n)" ] || (echo "Usage: make scale svc=<service> n=<count>"; exit 1)
	docker compose -f deployment/docker/docker-compose.yml up -d --scale $(svc)=$(n) --no-recreate

up:
	docker compose -f deployment/docker/docker-compose.yml up -d

down:
	docker compose -f deployment/docker/docker-compose.yml down