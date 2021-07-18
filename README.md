# Локальный jaeger + ядро микросервисов + примеры микросервисов

-----------------------------------------------

Сервисы управляются через конфиг **.env** и опции **make**. Поэтому первым делом запускаем **make help** или **make all**, изучаем

-----------------------------------------------

В проекте уже присутствуют три микросервиса в качестве примера взаимодействия, но надо **разово**, во-первых, создать для них конфиги и , во-вторых, создать схемы и объекты в базе:
* отредактировать в makefile опции базы
* Отработать **make create-env-%**, вместо % подставляем имя сервиса, изначально это **main**, **analytic** & **co**. Отредактировать **.env** каждого сервиса: уникальные - SERVICE_PORT, можно LOG_FILE_NAME.
* Отработать **make create-sql-%** Они создадут необходимые т-цы и триггера

-----------------------------------------------

Перед работой с сервисами всегда, в первую очередь, запускаем трассировщик - **make up-jaeger**

Сервисы запускаются в любом порядке - **make up-%**, вместо % подставляем имя сервиса

**Проверка работы**:
* **контейнеры** - *docker ps -a* После запуска должны присутствовать и работать 4 контейнера - *tpro_jaeger_1*, *main-1_tpro_nats_service_1*, *analytic-1_tpro_nats_service_1*, *co-1_tpro_nats_service_1*
* **логи** :
  * загрузочный конфиг - *docker logs main-1_tpro_nats_service_1*, *docker logs analytic-1_tpro_nats_service_1*, *docker logs co-1_tpro_nats_service_1*
  * logs/*.log
* **трассировщик** - *http://localhost:16686/search* или *http://jaeger.dev.lan:16686/search*. Хост трассировщика, должна быть картинка.

Грубая остановка сервиса с удалением образа - **make down-all-%**, вместо % подставляем имя сервиса

Аккуратная остановка сервиса - **make stop-%**, вместо % подставляем имя сервиса

-----------------------------------------------

**Cоздание нового сервиса**:

* Отрабатываем **make-create-%**, вместо % подставляем имя нового сервиса

Готово. Теперь по **make-up-%** должен подниматься новый сервис и слушать задания и шину. Вся бизнес-логика помещается в handlers_request.go & handlers_subscribe.go

-----------------------------------------------

<details><summary>Тестовые случаи</summary>

```
На всех операциях стоят паузы, поэтому записи в логах и трассировщики появляются не сразу. В трассировщике жмите кнопку Find Traces.

1. **Main** просит **analytic** сгенерить сид

INSERT INTO main_service.mq_publish (protocol,space_name,service_name,entity,method_name,timeout,headers,body) 
VALUES ('jsonrpc','tpro','analytic','example','simple',15,'{"request":{"type":"single"}, "content-language":"ru-RU"}','{"ip":"172.19.0.8", "login":"login342", "password":"phnd26w"}');

Смотрим трассировщик, логи. Возможны ошибки из-за неверного ip, логина, пароля.

2. **Main** просит **analytic** прислать сумму текущего часа + текущих минут. **analytic** делает это в параллельных потоках.

INSERT INTO main_service.mq_publish (protocol,space_name,service_name,entity,method_name,timeout,headers,body) 
VALUES ('jsonrpc','tpro','analytic','example','request_with_parallel_queries',15,'{"request":{"type":"single"}, "content-language":"ru-RU"}','{"ip":"172.19.0.8", "login":"login3422", "password":"phnd26w"}');

Смотрим трассировщик, логи. 

3. **Main** просит **analytic** прислать сумму текущего часа + текущих минут + (текущий час + 2) + (текущие минуты + 2). **analytic** делает это в параллельных потоках + посылает запрос в **со** и тот в свою очередь тоже распараллеливает расчет

INSERT INTO main_service.mq_publish (protocol,space_name,service_name,entity,method_name,timeout,headers,body)
VALUES ('jsonrpc','tpro','analytic','example','request_with_parallel_queries_with_transit_on_external_service',30,'{"request":{"type":"single"}, "content-language":"ru-RU"}','{"ip":"172.19.0.8", "login":"login3422", "password":"phnd26w"}');

Смотрим трассировщик, логи. 

```
</details>


## структура кода:

```
tpro_nats_microservices
├── core
│   ├── lib
│   │   ├── go.mod
│   │   ├── setup.go
│   │   └── utils.go
│   ├── newname
│   │   └── lib
│   │       ├── handlers_request.go
│   │       └── handlers_subscribe.go
│   ├── vendor
│   ├── docker-compose.yml
│   ├── dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── makefile
│   └── model_newname_service.sql
├── jaeger
│   ├── docker-compose.yml
│   ├── makefile
│   └── .env
├── logs
├── services
│   ├── analytic
│   │   ├── lib
│   │   │   ├── handlers_request.go
│   │   │   └── handlers_subscribe.go
│   │   └── .env
│   ├── co
│   │   ├── lib
│   │   │   ├── handlers_request.go
│   │   │   └── handlers_subscribe.go
│   │   └── .env
│   ├── main
│   │   ├── lib
│   │   │   ├── handlers_request.go
│   │   │   └── handlers_subscribe.go
│   │   └── .env
├── readme.md
├── makefile
└── .gitignore
```

## TODO

* [X] все опции (log, nats, jaeger) через конфиг (.env) 
* [X] вопросы хранения и версий либ (vendor & etc) - оставляем пока текущий вариант
* [X] поэкспериментировать с логом (в файл, ротация)
* [X] структура проекта - оставляем пока текущий вариант
* [X] health-checks nats & jaeger
* [X] token to nats
* [X] folder logs
* [X] make .env
* [X] graceful shutdown
* [ ] multi-stage build - временно отключил, кофликт с бесшовным обновлением, разберусь позже
* [X] бесшовное обновление сервиса, поддержка копий (см. ниже, примечание 1)
* [X] для каждого сервиса своя схема (временное решение)
* [X] синхронизация сообщений после падения какого-либо сервиса - отказ (см. ниже, примечание 2)
* [X] пример работы pwl-метода с микросервисом (#42359)
* [ ] анализ memory leaks
* [ ] тесты

### Примечание 1. Бесшовное обновление сервиса, поддержка копий:
* Чтобы локально создать копию сервиса надо в .env сервиса отредактировать COPY_NUMBER, SERVICE_PORT & LOG_FILE_NAME. После, запустить как обычно - make up-analytic
 Должен появиться контейнер, лог и вот здесь - http://nats_common.iac.tender.pro/subsz?subs=1, группа, например, "qgroup": "analytic-queue". Далее сам nats будет делать балансировку запросов к этой группе.

* Чтобы локально создать сервис другого релиза надо в .env сервиса отредактировать SERVICE_VERSION, COPY_NUMBER, SERVICE_PORT & LOG_FILE_NAME.
 Должен появиться контейнер, лог и вот здесь - http://nats_common.iac.tender.pro/subsz?subs=1, группа, например, "qgroup": "analytic-queue". Далее сам nats будет делать балансировку запросов к этой группе.

* Выкатка нового релиза выглядит так - старый сервис работает, запускаем сервис нового релиза, после того как он запущен даём команду старому корректно остановится - таким образом не будут потеряны ни задания для сервиса, ни собственные запросы к другим сервисам.

### Примечание 2. Синхронизация сообщений после падения какого-либо сервиса:
Отказ, т.к., во-первых, падение сервиса - это инцидент, не норма, и должен быть разбор причины и исправление; во-вторых, запуск нескольких копий поможет избежать потери сообщений; в-третьих, если сервис упал и сообщения не обработались, то можно легко, вручную повторить сообщения для этого сервиса, повторы отсеятся. Не вижу особого смысла сейчас автоматизировать обработку ситуации после падения сервиса, возможно в будущем. 


## вопросы девопс
* разворачивание nats
* jaeger
* elasticsearch? под jaeger & logs
