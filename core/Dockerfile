# STEP 1 build executable binary
FROM golang:1.14.0-alpine3.11 as builder

RUN mkdir -p /opt/app

WORKDIR /opt/app

COPY ./go.mod ./go.sum ./

COPY ./ ./

RUN GOOS=linux go build -ldflags="-s -w" -mod=vendor -v -o ./bin ./main.go

# закомментирован вариант 2-хшаговой компиляции, конфликт с бесшовным обновлением - разберусь позже
# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -mod=vendor -v -o ./bin ./main.go
# STEP 2 build a small image
# FROM alpine:3.4
# RUN apk --no-cache add ca-certificates
# COPY --from=builder /opt/app /opt/app

EXPOSE ${SERVICE_PORT}

ENTRYPOINT /opt/app/bin ./ --pguser ${DB_USER} --pghost ${DB_HOST} --pgport ${DB_PORT} --pgdbname ${DB_NAME} --pgpassword ${DB_PASSWORD} --url_nats ${URL_NATS} --service_name ${SERVICE_NAME} --pool_max ${DB_POOL_MAX} --service_version ${SERVICE_VERSION} --host_jaeger ${JAEGER_HOST} --port_jaeger ${JAEGER_PORT} --debug ${DEBUG} --log_file_name ${LOG_FILE_NAME} --max_size ${MAX_SIZE} --max_backups ${MAX_BACKUPS} --max_age ${MAX_AGE} --compress ${COMPRESS} --type_jaeger ${TYPE_JAEGER} --param_jaeger ${PARAM_JAEGER} --log_spans_jaeger ${LOG_SPANS_JAEGER} --total_wait_nats ${TOTAL_WAIT_NATS} --recon_delay_nats ${RECON_DELAY_NATS} --token_nats ${TOKEN_NATS}
