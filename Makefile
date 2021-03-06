SHELL = /bin/bash
CFG   ?= .env
DB_NAME            ?= tpro2_dev
DB_USER            ?= $(DB_NAME)
DCAPE_DB           ?= dcape_db_1
# ------------------------------------------------------------------------------
define CONFIG_DEF
# nats microservices config file, generated by make $(CFG)
SERVICE_NAME=newname
COPY_NUMBER=1
SERVICE_PORT=82
LOG_FILE_NAME=./logs/newname_service.log
SERVICE_VERSION=2.72
PROJECT_NAME=tpro_nats_microservice
DB_USER=$(DB_USER)
DB_HOST=db
DB_PORT=5432
DB_NAME=$(DB_NAME)
DB_PASSWORD=WBRwdfDXgmujAA
DB_POOL_MAX=6
URL_NATS=nats_auth.iac.tender.pro
JAEGER_HOST=jaeger
JAEGER_PORT=6831
DCAPE_NET=dcape_default
DC_VERSION=1.23.2
MODE=local
DEBUG=true
MAX_SIZE=1
MAX_BACKUPS=5
MAX_AGE=30
COMPRESS=false
TYPE_JAEGER=const
PARAM_JAEGER=1
LOG_SPANS_JAEGER=true
TOTAL_WAIT_NATS=9
RECON_DELAY_NATS=1
TOKEN_NATS=dag0HTXl4RGg7dXdaJwbC8
TIMEOUT_STOP=60
DCAPE_DB=$(DCAPE_DB)

########################################################################################################################################################################################################

# SERVICE_NAME     = название сервиса, должно быть уникально
# COPY_NUMBER      = 1, 2...Номер копии сервиса, можно запускать несколько одновременно
# SERVICE_PORT     = должен быть уникальным для каждого сервиса, иначе при запуске образа докера будет ошибка
# LOG_FILE_NAME    = Наименование и локация лога
# SERVICE_VERSION  = версия сервиса
# PROJECT_NAME     = оставьте как есть
# DB_USER          = юзер базы постгрес
# DB_HOST          = хост базы постгрес
# DB_PORT          = порт базы постгрес
# DB_NAME          = имя базы постгрес
# DB_PASSWORD      = пароль базы постгрес
# DB_POOL_MAX      = кол-во коннектов к базе, 0 или более 1 - если 0, то по дефолту будет равно NUMCPU 
# URL_NATS         = адрес шины, оставьте как есть
# JAEGER_HOST      = хост jaeger, оставьте как есть
# JAEGER_PORT      = порт jaeger, оставьте как есть
# DCAPE_NET        = сеть образов докера
# DC_VERSION       = версия докер-композ
# MODE             = среда (local, test, pre, rel)
# DEBUG            = Уровень вывода в лог, дебаг - true/false
# MAX_SIZE         = Макс.размер лога(мб)
# MAX_BACKUPS      = Кол-во бэкапов
# MAX_AGE          = Макс.число дней хранения бэкапов
# COMPRESS         = Бэкапы жмем? - true/false
# TYPE_JAEGER      = todo
# PARAM_JAEGER     = todo
# LOG_SPANS_JAEGER = todo
# TOTAL_WAIT_NATS  = todo
# RECON_DELAY_NATS = todo
# TOKEN_NATS       = mkpasswd (https://docs.nats.io/nats-server/configuration/securing_nats/auth_intro/tokens#bcrypted-tokens)
# TIMEOUT_STOP     = sec, timeout for graceful stop container. -> SIGTERM -> (timeout) -> SIGKILL
# DCAPE_DB         = контейнер базы
endef
# ------------------------------------------------------------------------------
-include $(CFG)
export
#-------------------------------------------------------------------------------
.PHONY: all help
# ------------------------------------------------------------------------------
create-%: ## Создать новый сервис %
create-%: y=$(subst create-,,$@)
create-%:
	@echo "************************* init service: $y"
	@echo "************************* create folder service: $y"
	cp -r core/newname* services/$y
	@echo "************************* create .env for service: $y"
	@[ -f services/$y/.env ] || { echo "$$CONFIG_DEF" > services/$y/.env ; echo "Warning: Created default services/$y/.env" ; }
	sed -i "s/newname/$y/" services/$y/.env
	@echo "************************* replace handlers service: $y"
	sed -i "s/newname/$y/" services/$y/lib/handlers_subscribe.go
	sed -i "s/newname/$y/" services/$y/lib/handlers_request.go
	@echo "************************* create postgres objects for service: $y"
	cp -r core/model_newname_service.sql init_newname_service.sql
	sed -i "s/newname/$y/" init_newname_service.sql
	cat init_newname_service.sql | docker exec -i $$DCAPE_DB psql -U $$DB_USER -d $$DB_NAME -1 -X;
	rm -f init_newname_service.sql
# ------------------------------------------------------------------------------
create-env-%: ## Создать .env для сервиса %
create-env-%: y=$(subst create-env-,,$@)
create-env-%:
	@echo "************************* create .env for service: $y"
	@[ -f services/$y/.env ] || { echo "$$CONFIG_DEF" > services/$y/.env ; echo "Warning: Created default services/$y/.env" ; }
	sed -i "s/newname/$y/" services/$y/.env
# ------------------------------------------------------------------------------
create-sql-%: ## Создать объекты базы для сервиса %
create-sql-%: y=$(subst create-sql-,,$@)
create-sql-%:
	@echo "************************* create postgres objects for service: $y"
	cp -r core/model_newname_service.sql init_newname_service.sql
	sed -i "s/newname/$y/" init_newname_service.sql
	cat init_newname_service.sql | docker exec -i $$DCAPE_DB psql -U $$DB_USER -d $$DB_NAME -1 -X;
	rm -f init_newname_service.sql
# ------------------------------------------------------------------------------
up-jaeger: ## Старт jaeger
	(cd jaeger && make up)
# ------------------------------------------------------------------------------
stop-jaeger: ## stop jaeger
	(cd jaeger && make stop)
# ------------------------------------------------------------------------------
down-all-jaeger: ## Стоп jaeger и удаление образа
	(cd jaeger && make down-all)
# ------------------------------------------------------------------------------
up-%: ## Старт сервиса %
up-%: y=$(subst up-,,$@)
up-%:
	@echo "************************* up service: $y"
	cp -r core/vendor services/$y/vendor
	cp -r core/lib/go.mod services/$y/lib/go.mod
	cp -r core/lib/setup.go services/$y/lib/setup.go
	cp -r core/lib/utils.go services/$y/lib/utils.go
	cp -r services/$y/lib services/$y/vendor/tpro/nats_service
	cp -r core/go.mod services/$y/go.mod
	cp -r core/go.sum services/$y/go.sum
	cp -r core/main.go services/$y/main.go
	cp -r core/Dockerfile services/$y/Dockerfile
	cp -r core/docker-compose.yml services/$y/docker-compose.yml
	cp -r core/Makefile services/$y/Makefile
	(cd services/$y && make up)
	rm -R services/$y/vendor || true
	rm -f services/$y/lib/go.mod || true
	rm -f services/$y/lib/setup.go || true
	rm -f services/$y/lib/utils.go || true
	rm -f services/$y/go.mod || true
	rm -f services/$y/go.sum || true
	rm -f services/$y/main.go || true

# ------------------------------------------------------------------------------
down-all-%: ## Стоп сервиса % и удаление образа
down-all-%: y=$(subst down-all-,,$@)
down-all-%:
	@echo "************************* down-all service: $y"
	cp -r core/Dockerfile services/$y/Dockerfile
	cp -r core/docker-compose.yml services/$y/docker-compose.yml
	cp -r core/Makefile services/$y/Makefile
	(cd services/$y && make down-all)
	rm -f services/$y/Dockerfile
	rm -f services/$y/docker-compose.yml
	rm -f services/$y/Makefile
# ------------------------------------------------------------------------------
stop-%: ## Аккуратный стоп сервиса %
stop-%: y=$(subst stop-,,$@)
stop-%:
	@echo "************************* graceful stop service: $y"
	cp -r core/Dockerfile services/$y/Dockerfile
	cp -r core/docker-compose.yml services/$y/docker-compose.yml
	cp -r core/Makefile services/$y/Makefile
	(cd services/$y && make stop)
	rm -f services/$y/Dockerfile
	rm -f services/$y/docker-compose.yml
	rm -f services/$y/Makefile
# ------------------------------------------------------------------------------
all: ## Все опции make
all: help
# ------------------------------------------------------------------------------
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-%]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
