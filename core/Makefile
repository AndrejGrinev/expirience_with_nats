SHELL = /bin/bash
CFG   ?= .env
# ------------------------------------------------------------------------------
-include $(CFG)
export
#-------------------------------------------------------------------------------
.PHONY: all help down clean-docker down-all reup up
# ------------------------------------------------------------------------------
all: ## Full list options
all: help
# ------------------------------------------------------------------------------
up: ## Start pg & app containers
up: CMD=up -d
up: dc
# ------------------------------------------------------------------------------
reup: ## Restart pg & app containers
reup: CMD=up --force-recreate -d
reup: dc
# ------------------------------------------------------------------------------
down: ## Stop containers and remove them
down: CMD=rm -f -s
down: dc
# ------------------------------------------------------------------------------
clean-docker: ## Delete docker image
		docker rmi ${SERVICE_NAME}_${PROJECT_NAME}:$(MODE)-$(SERVICE_VERSION)
# ------------------------------------------------------------------------------
down-all: ## Stop containers and remove them & delete docker image
down-all: down clean-docker
# ------------------------------------------------------------------------------
stop: ## graceful stop container
stop: CMD=stop --time=${TIMEOUT_STOP}
stop: dc
# ------------------------------------------------------------------------------
# $$PWD используется для того, чтобы текущий каталог был доступен в контейнере по тому же пути
# и относительные тома новых контейнеров могли его использовать
## run docker-compose
dc: docker-compose.yml
	@docker run --rm  \
	  -v /var/run/docker.sock:/var/run/docker.sock \
	  -v $(PWD):$(PWD) \
	  -w $(PWD) \
	  docker/compose:$(DC_VERSION) \
	  -p $(SERVICE_NAME)-$(COPY_NUMBER) \
	  $(CMD)
# ------------------------------------------------------------------------------
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
