module tpro/nats_service

go 1.13

replace tpro/nats_service/lib => ./lib

require (
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/jackc/pgx/v4 v4.5.0 // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/nats-io/nats-server/v2 v2.1.9 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	go.uber.org/zap v1.10.0
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	tpro/nats_service/lib v0.0.0-00010101000000-000000000000
)
