package lib

import (
	"context"
	"go.uber.org/zap"
	"time"
	jaeger "github.com/uber/jaeger-client-go"
)
// доступные вызовы сервиса, внимание!!! вместо *newname* здесь надо вписывать название нужного сервиса
var (
	handler_1 string = "jsonrpc://tpro.newname.company.winners"
)
// Хеш-карта доступных вызовов сервиса и его обработчиков
var MapHandlersFuncSubs = map[string]func(msg *RequestMessage, logger *zap.Logger) (map[string]interface{}, error) {
	handler_1: company_winners,
}
// Роутер обработчиков
func HubFuncS(msg *RequestMessage, f func(*RequestMessage, *zap.Logger) (map[string]interface{}, error), logger *zap.Logger) (map[string]interface{}, error) {
	return f(msg, logger)
}
// Обработчик
func company_winners(msg *RequestMessage, logger *zap.Logger) (map[string]interface{}, error) {
	span := createSpan ("exampleSimple", "exampleSimple", msg.Headers, logger)
	sctx, okk := span.Context().(jaeger.SpanContext)
	if okk {
	        msg.Headers["trace-id"] = sctx.String()
	}
	defer span.Finish()
	logger.Debug("[" + Cfg.ServiceName + ".company_winners] on job(7)")
	time.Sleep(7 * time.Second)
	var res int
	conn, err := Db_sub.Acquire(context.Background())
	if err != nil {
		logger.Fatal("[" + Cfg.ServiceName + ".company_winners] " + "Error acquiring connection:", zap.Error(err))
	}
	defer conn.Release()
	err_query := conn.QueryRow(context.Background(), "SELECT date_part('minute',now())").Scan(&res)
	if err_query != nil {
		logger.Error("[" + Cfg.ServiceName + ".company_winners] " + "Error query:", zap.Error(err_query))
	}
	span.LogKV("result", res)
	return map[string]interface{}{
		"winners":     res,
	}, nil
}
