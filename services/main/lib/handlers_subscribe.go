package lib

import (
	"context"
	"go.uber.org/zap"
	"time"
	jaeger "github.com/uber/jaeger-client-go"
)
// доступные вызовы сервиса, внимание!!! вместо *main* здесь надо вписывать название нужного сервиса
var (
	handler_1 string = "jsonrpc://tpro.main.company.winners"
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
	span := createSpan ("company_winners", "company_winners", msg.Headers, logger)
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
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "company_winners"), zap.Any("context", "Can't assert mq_id, handle error."))
			mq_id = 1
		}
	} else {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "company_winners"), zap.Any("context", "not found mq_id, handle error."))
		mq_id = 0
	}
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "company_winners"), zap.Any("context", "on job(7)"), zap.Int("reply_mq_id", mq_id))
	time.Sleep(7 * time.Second)
	var res int
	conn, err := Db_sub.Acquire(context.Background())
	if err != nil {
		logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "company_winners"), zap.Any("context", "Error acquiring connection"), zap.Error(err))
	}
	defer conn.Release()
	err_query := conn.QueryRow(context.Background(), "SELECT date_part('minute',now())").Scan(&res)
	if err_query != nil {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "company_winners"), zap.Any("context", "Error query"), zap.Error(err_query))
	}
	span.LogKV("result", res)
	return map[string]interface{}{
		"winners":     res,
	}, nil
}
