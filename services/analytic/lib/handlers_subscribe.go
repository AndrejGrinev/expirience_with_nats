package lib

import (
	"context"
	"go.uber.org/zap"
	"fmt"
	"time"
	opentracing "github.com/opentracing/opentracing-go"
	"errors"
	jaeger "github.com/uber/jaeger-client-go"
)
var (
	handler_1 string = "jsonrpc://tpro.analytic.example.simple"
	handler_2 string = "jsonrpc://tpro.analytic.example.request_with_parallel_queries"
	handler_3 string = "jsonrpc://tpro.analytic.example.request_with_parallel_queries_with_transit_on_external_service"
)
// Хеш-карта доступных вызовов сервиса и его обработчиков
var MapHandlersFuncSubs = map[string]func(msg *RequestMessage, logger *zap.Logger) (map[string]interface{}, error) {
	handler_1: exampleSimple,
	handler_2: exampleRequestWithParallelQueries,
	handler_3: exampleRequestWithParallelQueriesWithTransitOnExternalService,
}
// Роутер обработчиков
func HubFuncS(msg *RequestMessage, f func(*RequestMessage, *zap.Logger) (map[string]interface{}, error), logger *zap.Logger) (map[string]interface{}, error) {
	return f(msg, logger)
}
// Простой пример, просто запрос - просто данные
func exampleSimple(msg *RequestMessage, logger *zap.Logger) (map[string]interface{}, error) {
	span := createSpan ("exampleSimple", "exampleSimple", msg.Headers, logger)
	sctx, okk := span.Context().(jaeger.SpanContext)
	if okk {
	        msg.Headers["trace-id"] = sctx.String()
	}
	defer span.Finish()
	var mq_id int
	if x, found := msg.Headers["mq_id"]; found {
		if mqq_id, ok := x.(float64); ok {
			mq_id = int(mqq_id)
		} else {
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "Can't assert mq_id, handle error."))
			mq_id = 1
		}
	} else {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "not found mq_id, handle error."))
		mq_id = 0
	}
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "on job(2)"), zap.Int("reply_mq_id", mq_id))
//	time.Sleep(2 * time.Second)
	var sid, face_id int
	var ip, login, password string
	var ok bool
	if x, found := msg.Body["ip"]; found {
		if ip, ok = x.(string); !ok {
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "ip found, but not string"), zap.Int("reply_mq_id", mq_id))
			span.SetTag("error", true)
			return nil, errors.New("ip found, but not string")
		}
	} else {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "ip not found"), zap.Int("reply_mq_id", mq_id))
		span.SetTag("error", true)
		return nil, errors.New("ip not found")
	}
	if x, found := msg.Body["login"]; found {
		if login, ok = x.(string); !ok {
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "login found, but not string"), zap.Int("reply_mq_id", mq_id))
			span.SetTag("error", true)
			return nil, errors.New("login found, but not string")
		}
	} else {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "login not found"), zap.Int("reply_mq_id", mq_id))
		span.SetTag("error", true)
		return nil, errors.New("login not found")
	}
	if x, found := msg.Body["password"]; found {
		if password, ok = x.(string); !ok {
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "password found, but not string"), zap.Int("reply_mq_id", mq_id))
			span.SetTag("error", true)
			return nil, errors.New("password found, but not string")
		}
	} else {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "password not found"), zap.Int("reply_mq_id", mq_id))
		span.SetTag("error", true)
		return nil, errors.New("password not found")
	}
	conn, err := Db_sub.Acquire(context.Background())
	if err != nil {
		span.SetTag("error", true)
		logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "Error acquiring connection"), zap.Int("reply_mq_id", mq_id), zap.Error(err))
	}
	defer conn.Release()
	err_query := conn.QueryRow(context.Background(), "SELECT sid, face_id from app.login($1, $2, $3)", ip, login, password).Scan(&sid, &face_id)
	if err_query != nil {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleSimple"), zap.Any("context", "Error query"), zap.Int("reply_mq_id", mq_id), zap.Error(err_query))
		span.SetTag("error", true)
		return nil, err_query
	}
	span.LogKV("result", map[string]interface{}{"sid":sid,"face_id":face_id,})
	return map[string]interface{}{
		"sid":     sid,
		"face_id": face_id,
	}, nil
}
// Пример, распараллеливание на несколько запросов внутри сервиса без транзита в другой сервис
func exampleRequestWithParallelQueries(msg *RequestMessage, logger *zap.Logger) (map[string]interface{}, error) {
	span := createSpan ("exampleRequestWithParallelQueries", "exampleRequestWithParallelQueries", msg.Headers, logger)
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
	span.LogKV("result", res1 + res2)
	return map[string]interface{}{
		"exampleRequestWithParallelQueries": res1 + res2,
	}, nil
}
// Пример, распараллеливание на несколько запросов внутри сервиса + транзит в другой сервис
func exampleRequestWithParallelQueriesWithTransitOnExternalService(msg *RequestMessage, logger *zap.Logger) (map[string]interface{}, error) {
	span := createSpan ("exampleRequestWithParallelQueriesWithTransitOnExternalService", "exampleRequestWithParallelQueriesWithTransitOnExternalService", msg.Headers, logger)
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
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleRequestWithParallelQueriesWithTransitOnExternalService"), zap.Any("context", "Can't assert mq_id, handle error."))
			mq_id = 1
		}
	} else {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleRequestWithParallelQueriesWithTransitOnExternalService"), zap.Any("context", "not found mq_id, handle error."))
		mq_id = 0
	}
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleRequestWithParallelQueriesWithTransitOnExternalService"), zap.Any("context", "on job(3)"), zap.Int("reply_mq_id", mq_id))
	time.Sleep(3 * time.Second)
	c1 := make(chan int)
	c2 := make(chan int)
	chanMessage := &ChanMessage{}
	chanMessage.Body = make(chan map[string]interface{})
	chanMessage.Transit = make(chan bool)
	go parallelQuery1(span, logger, c1)
	go parallelQuery2(span, logger, c2)
	go transitService(span, logger, chanMessage)
	res1 := <-c1
	res2 := <-c2
	res3 := <-chanMessage.Body
	v, ok := res3["result"].(map[string]interface{})
	if !ok {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "exampleRequestWithParallelQueriesWithTransitOnExternalService"), zap.Any("context", "Can't assert, handle error."))
		span.SetTag("error", true)
	}
	var res3_parse int
	res3_parse = int(v["exampleRequestWithParallelQueries"].(float64))
	close(c1)
	close(c2)
	close(chanMessage.Body)
	var res int
	res = res1 + res2 + res3_parse
	span.LogKV("result", res)
	return map[string]interface{}{
		"exampleRequestWithParallelQueriesWithTransitOnExternalService": res,
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
	err_query := conn.QueryRow(context.Background(), "SELECT date_part('minute',now())").Scan(&res)
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
	err_query := conn.QueryRow(context.Background(), "SELECT date_part('hour',now())").Scan(&res)
	if err_query != nil {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "parallelQuery2"), zap.Any("context", "Error query"), zap.Error(err_query))
		span.SetTag("error", true)
	}
	span.LogKV("result", res)
	c <- res
}
// Транзитный вызов
func transitService(span opentracing.Span, logger *zap.Logger, c *ChanMessage) {
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "transitService"), zap.Any("context", "on job(7)"))
	span = opentracing.StartSpan("transitService", opentracing.ChildOf(span.Context()))
	defer span.Finish()
	sctx, _ := span.Context().(jaeger.SpanContext)
	headers := map[string]string{"trace-id": sctx.String()}
	time.Sleep(7 * time.Second)
	conn, err := Db_sub.Acquire(context.Background())
	if err != nil {
		logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "transitService"), zap.Any("context", "Error acquiring connection"), zap.Error(err))
		span.SetTag("error", true)
	}
	defer conn.Release()
	var res int
	for {
		err_query := conn.QueryRow(context.Background(), "insert into analytic_service.mq_publish(protocol,space_name,service_name,entity,method_name,timeout,headers,body, direct) values ('jsonrpc','tpro','co','example','request_with_parallel_queries',20,$1,'{}', true) returning id", &headers).Scan(&res)
		if err_query != nil {
			logger.Warn("", zap.String("service", Cfg.ServiceName), zap.String("function", "transitService"), zap.Any("context", "Error query"), zap.Error(err_query))
		} else {
			break
		}
		time.Sleep(time.Millisecond)
	}
	span.LogKV("direct id", res)
	Wg_req.Add(1)
	// счетчик активных нотифаев
	CountNotify++
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "transitService"), zap.Any("context", "how many active notify?"), zap.Any("CountNotify", CountNotify))
	go haveMessage (context.Background(), fmt.Sprint(res), logger, c)
	<-c.Transit
	close(c.Transit)
}
