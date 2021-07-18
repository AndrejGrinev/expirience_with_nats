package lib

import (
	"go.uber.org/zap"
//	"strconv"
)
var (
	handler_11 string = "jsonrpc://tpro.analytic.example.simple"
	handler_22 string = "jsonrpc://tpro.analytic.example.request_with_parallel_queries"
	handler_33 string = "jsonrpc://tpro.analytic.example.request_with_parallel_queries_with_transit_on_external_service"
)
// Хеш-карта доступных вызовов сервиса и обработчиков положительного ответа
var MapHandlersFuncReply = map[string]func(msg map[string]interface{}, logger *zap.Logger, reply_to string) int {
	handler_11: analyticExampleSimple,
	handler_22: analyticExampleRequestWithParallelQueries,
	handler_33: analyticExampleRequestWithParallelQueriesWithTransitOnexternalService,
}
// Хеш-карта доступных вызовов сервиса и обработчиков негативного ответа
var MapHandlersFuncReplyError = map[string]func(logger *zap.Logger) int {
	handler_11: analyticExampleSimpleError,
	handler_22: analyticExampleRequestWithParallelQueriesError,
	handler_33: analyticExampleRequestWithParallelQueriesWithTransitOnexternalServiceError,
}
// Роутер обработчиков положительного ответа
func HubFunc(msg map[string]interface{}, f func(map[string]interface{}, *zap.Logger, string) int, logger *zap.Logger, reply_to string) int {
	return f(msg, logger, reply_to)
}
// Роутер обработчиков негативного ответа
func HubFuncError(f func(*zap.Logger) int, logger *zap.Logger) int {
	return f(logger)
}
// Обработка положительного ответа
func analyticExampleSimple(msg map[string]interface{}, logger *zap.Logger, reply_to string) int {
	v, ok := msg["meta"].(map[string]interface{})
	if !ok {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleSimple"), zap.Any("context", "Can't assert, handle error."))
	}
	var mq_id int
	mq_id = int(v["mq_id"].(float64))
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleSimple"), zap.Any("context", "Received data"), zap.Int("request_mq_id", mq_id), zap.Reflect("msg.result", msg["result"]), zap.Reflect("msg.error", msg["error"]))
	subj := "$$.nats-pwl-reply." + reply_to
	if msg["error"] != nil {
		_ = Ec.Publish(subj, msg["error"])
	} else {
		_ = Ec.Publish(subj, msg["result"])
	}
	return 1
}
// Обработка негативного ответа
func analyticExampleSimpleError(logger *zap.Logger) int {
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleSimpleError"), zap.Any("context", "Received error"))
	return 2
}
// Обработка положительного ответа
func analyticExampleRequestWithParallelQueries(msg map[string]interface{}, logger *zap.Logger, reply_to string) int {
	v, ok := msg["meta"].(map[string]interface{})
	if !ok {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleRequestWithParallelQueries"), zap.Any("context", "Can't assert, handle error."))
	}
	var mq_id int
	mq_id = int(v["mq_id"].(float64))
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleRequestWithParallelQueries"), zap.Any("context", "Received data"), zap.Int("request_mq_id", mq_id), zap.Reflect("msg.result", msg))
	return 3
}
// Обработка негативного ответа
func analyticExampleRequestWithParallelQueriesError(logger *zap.Logger) int {
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleRequestWithParallelQueriesError"), zap.Any("context", "Received error"))
	return 4
}
// Обработка положительного ответа
func analyticExampleRequestWithParallelQueriesWithTransitOnexternalService(msg map[string]interface{}, logger *zap.Logger, reply_to string) int {
	v, ok := msg["meta"].(map[string]interface{})
	if !ok {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleRequestWithParallelQueriesWithTransitOnexternalService"), zap.Any("context", "Can't assert, handle error."))
	}
	var mq_id int
	mq_id = int(v["mq_id"].(float64))
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleRequestWithParallelQueriesWithTransitOnexternalService"), zap.Any("context", "Received data"), zap.Int("request_mq_id", mq_id), zap.Reflect("msg.result", msg))
	return 5
}
// Обработка негативного ответа
func analyticExampleRequestWithParallelQueriesWithTransitOnexternalServiceError(logger *zap.Logger) int {
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "analyticExampleRequestWithParallelQueriesWithTransitOnexternalServiceError"), zap.Any("context", "Received error"))
	return 6
}
