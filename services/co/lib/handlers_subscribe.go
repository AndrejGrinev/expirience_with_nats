package lib

import (
	"context"
	"go.uber.org/zap"
	"time"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
)
// доступные вызовы сервиса, внимание!!! вместо *co* здесь надо вписывать название нужного сервиса
var (
	handler_1 string = "jsonrpc://tpro.co.example.request_with_parallel_queries"
)
// Хеш-карта доступных вызовов сервиса и его обработчиков
var MapHandlersFuncSubs = map[string]func(msg *RequestMessage, logger *zap.Logger) (map[string]interface{}, error) {
	handler_1: exampleRequestWithParallelQueries,
}
// Роутер обработчиков
func HubFuncS(msg *RequestMessage, f func(*RequestMessage, *zap.Logger) (map[string]interface{}, error), logger *zap.Logger) (map[string]interface{}, error) {
	return f(msg, logger)
}
// Пример, распараллеливание на несколько запросов внутри сервиса без транзита в другой сервис
func exampleRequestWithParallelQueries(msg *RequestMessage, logger *zap.Logger) (map[string]interface{}, error) {
	span := createSpan ("exampleRequestWithParallelQueries", "exampleRequestWithParallelQueries", msg.Headers, logger)
	span.LogKV("header", msg.Headers)
	span.LogKV("body", msg.Body)
	sctx, ok := span.Context().(jaeger.SpanContext)
	if ok {
	        msg.Headers["trace-id"] = sctx.String()
	}
	defer span.Finish()
	var mq_id int
	if x, found := msg.Headers["mq_id"]; found {
		if mqq_id, ok := x.(float64); ok {
			mq_id = int(mqq_id)
		} else {
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleRequestWithParallelQueries"), zap.Any("context", "Can't assert mq_id, handle error."))
			mq_id = 1
		}
	} else {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleRequestWithParallelQueries"), zap.Any("context", "not found mq_id, handle error."))
		mq_id = 0
	}
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleRequestWithParallelQueries"), zap.Any("context", "on job(4)"), zap.Int("reply_mq_id", mq_id))
	time.Sleep(4 * time.Second)
	c1 := make(chan int)
	c2 := make(chan int)
	go parallelQuery1(span, logger, c1)
	go parallelQuery2(span, logger, c2)
	res1 := <-c1
	res2 := <-c2
	close(c1)
	close(c2)
	var res int
	res = res1 + res2
	span.LogKV("result", res)
	return map[string]interface{}{
		"exampleRequestWithParallelQueries": res,
	}, nil
}
// Первая параллель
func parallelQuery1(span opentracing.Span, logger *zap.Logger, c chan int) {
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "parallelQuery1"), zap.Any("context", "on job(3)"))
	span = opentracing.StartSpan("parallelQuery1", opentracing.ChildOf(span.Context()))
	defer span.Finish()
	time.Sleep(3 * time.Second)
	conn, err := Db_sub.Acquire(context.Background())
	if err != nil {
		logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "parallelQuery1"), zap.Any("context", "Error acquiring connection"), zap.Error(err))
		span.SetTag("error", true)
	}
	defer conn.Release()
	var res int
	err_query := conn.QueryRow(context.Background(), "SELECT date_part('minute',now()) + 2").Scan(&res)
	if err_query != nil {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "parallelQuery1"), zap.Any("context", "Error query"), zap.Error(err_query))
		span.SetTag("error", true)
	}
	span.LogKV("result", res)
	c <- res
}
// Вторая параллель
func parallelQuery2(span opentracing.Span, logger *zap.Logger, c chan int) {
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "parallelQuery2"), zap.Any("context", "on job(5)"))
	span = opentracing.StartSpan("parallelQuery2", opentracing.ChildOf(span.Context()))
	defer span.Finish()
	time.Sleep(5 * time.Second)
	conn, err := Db_sub.Acquire(context.Background())
	if err != nil {
		logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "parallelQuery2"), zap.Any("context", "Error acquiring connection"), zap.Error(err))
		span.SetTag("error", true)
	}
	defer conn.Release()
	var res int
	err_query := conn.QueryRow(context.Background(), "SELECT date_part('hour',now()) + 2").Scan(&res)
	if err_query != nil {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "parallelQuery2"), zap.Any("context", "Error query"), zap.Error(err_query))
		span.SetTag("error", true)
	}
	span.LogKV("result", res)
	c <- res
}
