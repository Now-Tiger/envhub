# ============================================================
# 🚀 EnvHub • Docker Command Centre
# ============================================================
.PHONY: help build up down restart logs logs-db logs-all \
        status health shell psql db-status clean purge

COMPOSE  := docker compose -f docker-compose.yml
API_URL  := http://localhost:8080

R  := \033[0m
B  := \033[1m
DM := \033[2m
CY := \033[36m
GN := \033[32m
YL := \033[33m
RD := \033[31m
WH := \033[97m
MG := \033[35m

line = @printf "$(DM)────────────────────────────────────────────────────────────$(R)\n"
bigline = @printf "$(B)$(CY)════════════════════════════════════════════════════$(R)\n"

step = @printf "\n$(B)$(CY)🚀  %-35s$(R)$(DM) %s$(R)\n" $(1) $(2)
ok   = @printf "$(GN)✔ SUCCESS$(R)  %s\n" $(1)
info = @printf "$(CY)ℹ INFO   $(R)  %s\n" $(1)
warn = @printf "$(YL)⚠ WARN   $(R)  %s\n" $(1)
fail = @printf "$(RD)✘ ERROR  $(R)  %s\n" $(1)

.DEFAULT_GOAL := help

help:
	@printf "\n"
	$(call bigline)
	@printf "$(B)$(WH)🚀 EnvHub Docker Command Centre$(R)\n"
	$(call bigline)
	@printf "\n$(B)$(CY)COMMAND$(R)               $(B)DESCRIPTION$(R)\n"
	$(call line)
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
	awk 'BEGIN { FS = ":.*?##" } { \
	    split($$1, cmd, ":"); \
	    printf "  \033[36m%-20s\033[0m  \033[37m%s\033[0m\n", cmd[1], $$2 \
	}'
	@printf "\n$(DM)Example → make up | make logs | make psql$(R)\n\n"

build: ## Build Docker images
	$(call bigline)
	$(call step,"Building Docker images...","this may take a while ⏳")
	$(call line)
	@$(COMPOSE) build
	$(call line)
	$(call ok,"All images built successfully 🎉")

up: ## Start full stack
	$(call bigline)
	$(call step,"Starting EnvHub stack...","postgres → api")
	$(call line)
	@$(COMPOSE) up --build -d
	$(call line)
	$(call ok,"Stack is LIVE 🚀")
	$(call info,"API  →  $(API_URL)")
	$(call line)

down: ## Stop stack
	$(call bigline)
	$(call step,"Stopping services...","gracefully shutting down")
	$(call line)
	@$(COMPOSE) down --remove-orphans
	$(call ok,"Containers stopped 🛑 (data preserved)")

restart: ## Restart API
	$(call bigline)
	$(call step,"Restarting API container...","")
	$(call line)
	@$(COMPOSE) restart api
	$(call ok,"API restarted 🔄")

ps: ## list containers
	$(call step,"Running containers","")
	$(call line)
	@$(COMPOSE) ps

logs: ## API logs
	$(call bigline)
	$(call step,"Streaming API logs...","Ctrl+C to exit")
	$(call line)
	@$(COMPOSE) logs -f --tail=100 api

logs-db: ## DB logs
	$(call bigline)
	$(call step,"Streaming PostgreSQL logs...","Ctrl+C to exit")
	$(call line)
	@$(COMPOSE) logs -f --tail=100 postgres

logs-all: ## All logs
	$(call bigline)
	$(call step,"Streaming ALL logs...","Ctrl+C to exit")
	$(call line)
	@$(COMPOSE) logs -f --tail=50

status: ## Show container status
	$(call bigline)
	$(call step,"Checking container status...","")
	$(call line)
	@$(COMPOSE) ps

health: ## API health check
	$(call bigline)
	$(call step,"Checking API health...","")
	$(call line)
	@curl -s --max-time 5 $(API_URL)/health \
	    | python3 -m json.tool 2>/dev/null \
	    || ($(call fail,"API not responding ❌ Try: make up") && exit 1)

psql: ## Open DB shell
	$(call bigline)
	$(call step,"Opening PostgreSQL shell...","type \\q to exit")
	$(call line)
	@$(COMPOSE) exec postgres psql -U envhub_user -d envhub_db 

db-status: ## DB table stats
	$(call bigline)
	$(call step,"Fetching DB table stats...","")
	$(call line)
	@$(COMPOSE) exec -T postgres psql -U envhub_user -d envhub_db \
	    -c "SELECT tablename AS table, n_live_tup AS rows FROM pg_stat_user_tables ORDER BY tablename;"

shell: ## API shell
	$(call bigline)
	$(call step,"Opening API shell...","type exit to leave")
	$(call line)
	@$(COMPOSE) exec api /bin/bash

clean: ## Remove containers (keep data)
	$(call bigline)
	$(call step,"Cleaning containers...","data is safe")
	$(call line)
	@$(COMPOSE) down --remove-orphans
	$(call ok,"Cleanup complete 🧹")

purge: ## Destroy everything
	$(call bigline)
	$(call warn,"⚠ THIS WILL DELETE ALL DATA PERMANENTLY")
	@printf "$(B)$(RD)Are you sure? [y/N]: $(R)"; read ans; \
	[ "$$ans" = "y" ] || { echo "Aborted."; exit 1; }
	$(call line)
	@$(COMPOSE) down -v --remove-orphans
	$(call ok,"Everything wiped 💀")
