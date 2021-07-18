package lib

import (
	"go.uber.org/zap"
)
var (
	handler_11 string = "jsonrpc://tpro.co.example.request_with_parallel_queries"
)
// Хеш-карта доступных вызовов сервиса и обработчиков положительного ответа
var MapHandlersFuncReply = map[string]func(msg map[string]interface{}, logger *zap.Logger, reply_to string) int {
	handler_11: coExampleRequestWithParallelQueries,
}
// Хеш-карта доступных вызовов сервиса и обработчиков негативного ответа
var MapHandlersFuncReplyError = map[string]func(logger *zap.Logger) int {
	handler_11: coExampleRequestWithParallelQueriesError,
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
func coExampleRequestWithParallelQueries(msg map[string]interface{}, logger *zap.Logger, reply_to string) int {
	v, ok := msg["meta"].(map[string]interface{})
	if !ok {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "coExampleRequestWithParallelQueries"), zap.Any("context", "Can't assert, handle error."))
	}
	var mq_id int
	mq_id = int(v["mq_id"].(float64))
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "coExampleRequestWithParallelQueries"), zap.Any("context", "Received data"), zap.Int("request_mq_id", mq_id), zap.Reflect("msg.result", msg))
	return 1
}
// Обработка негативного ответа
func coExampleRequestWithParallelQueriesError(logger *zap.Logger) int {
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "coExampleRequestWithParallelQueriesError"), zap.Any("context", "Received error"))
	return 2
}
