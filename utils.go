package lib

import (
	"context"
	"time"
	"go.uber.org/zap"
	"fmt"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
)
// Запуск слушателя нотифаев
func listen(ctx context.Context, logger *zap.Logger, tracer opentracing.Tracer) {
	// берём коннект из пула
	conn, err := Db_req.Acquire(ctx)
	if err != nil {
		logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "listen"), zap.Any("context", "Error acquiring connection"), zap.Error(err))
	}
	defer conn.Release()
	// начинаем слушать канал нотифай
        listenNotifyCnannel := "listen task_for_" + Cfg.ServiceName + "_service"
	for {
		_, err := conn.Exec(ctx, listenNotifyCnannel)
		if err != nil {
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "listen"), zap.Any("context", "Error listening to notify channel"), zap.Error(err))
		} else {
			break
		}
		time.Sleep(time.Second)
	}
	logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "listen"), zap.Any("context", "Start listen notification channel"), zap.String("channel", "task_for_" + Cfg.ServiceName + "_service"))
	// бесконечный цикл
	for {
		// ждём нотифай из канала
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if Stopped {
				break
			} else {
				logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "listen"), zap.Any("context", "Error waiting for notification"), zap.Error(err))
			}
		} else {
			logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "listen"), zap.Any("context", "Get notification"), zap.Uint32("PID", notification.PID), zap.String("Payload", notification.Payload))
			// что-то получили
			if !Stopped {
				Wg_req.Add(1)
				// счетчик активных нотифаев
				CountNotify++
				chanMessage := &ChanMessage{}
				go haveMessage (ctx, notification.Payload, logger, chanMessage)
			}
			logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "listen"), zap.Any("context", "How many active notify?"), zap.Any("CountNotify", CountNotify))
		}
	}
}
// Обработчик нотифаев
func haveMessage (ctx context.Context, notification_id string, logger *zap.Logger, c *ChanMessage) {
	defer Wg_req.Done()
	var mq_id, timeout int
        var service_name, target, reply_to string
	var headers, body map[string]interface{}
	// берём коннект из пула
	conn, err := Db_req.Acquire(ctx)
	if err != nil {
		logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "Error acquiring connection"), zap.Error(err))
	}
	defer conn.Release()
	err_query := conn.QueryRow(ctx, "SELECT mq_id, service_name, protocol || '://' || space_name || '.' || service_name || '.' || entity || '.' || method_name, timeout, headers, body, reply_to FROM " + Cfg.ServiceName + "_service.mq_publish WHERE id = $1", notification_id).Scan(&mq_id, &service_name, &target, &timeout, &headers, &body, &reply_to)
	if err_query != nil {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "Error query"), zap.Error(err_query))
	}
	logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "notification data"), zap.Int("request_mq_id", mq_id), zap.String("service_name", service_name), zap.String("target", target), zap.Int("timeout", timeout), zap.Any("headers", headers), zap.Any("body", body))
	headers["mq_id"] = int(mq_id)
	headers["from_service"] = Cfg.ServiceName
	span := createSpan (target, "haveMessage", headers, logger)
	span.LogKV("notification_id", notification_id)
	sctx, ok := span.Context().(jaeger.SpanContext)
	if ok {
	        headers["trace-id"] = sctx.String()
	}
	defer span.Finish()
	// ветка с timeout == 0 пока не перепроверена, хотя работала до внедрения спанов
        payload := []byte(notification_id)
	if timeout == 0 {
		// работает PUBLISH
		Ec.Publish(target, payload)
		Ec.Flush()
		if err := Ec.LastError(); err != nil {
			// случилась ошибка - время вышло или ещё что-то
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "Error publish"), zap.Int("request_mq_id", mq_id), zap.Error(Ec.LastError()))
		} else {
			// success
			logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "Published"), zap.Int("request_mq_id", mq_id), zap.String(target, string(payload)))
		}
		// считаем, что отработал нотифай 
		CountNotify--
	} else {
		// делаем реквест в NATS
		var response map[string]interface{}
		err := Ec.Request(target, &RequestMessage{Headers: headers, Body: body}, &response, time.Duration(timeout)*time.Second)
		logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "Get answer from"), zap.Int("request_mq_id", mq_id), zap.String("target", target))
		span.LogKV("result", response)
		if c.Body == nil {
			logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "not have chan"), zap.Int("request_mq_id", mq_id))
		} else {
			logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "has chan"), zap.Int("request_mq_id", mq_id))
			c.Body <- response
			c.Transit <- true
		}
		if err != nil {
			// случилась ошибка - время вышло или ещё что-то
			if Ec.LastError() != nil {
				logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "Error request"), zap.Int("request_mq_id", mq_id), zap.Error(Ec.LastError()))
			}
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "Error request"), zap.Int("request_mq_id", mq_id), zap.Error(err))
			HubFuncError(MapHandlersFuncReplyError[target], logger)
		} else {
			if response == nil {
				logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "No response received"), zap.Int("request_mq_id", mq_id))
			} else {
				// ответ получен
				if _, ok := MapHandlersFuncReply[target]; ok {
					HubFunc(response, MapHandlersFuncReply[target], logger, reply_to)
				} else {
					span.SetTag("error", true)
					logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "wrong handler"), zap.Int("request_mq_id", mq_id))
				}
			}
		}
		// считаем, что отработал нотифай 
		CountNotify--
		logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "haveMessage"), zap.Any("context", "How many active notify?"), zap.Int("request_mq_id", mq_id), zap.Any("CountNotify", CountNotify))
	}
}
// Запуск слушателя других сервисов
func Service(ctx context.Context, logger *zap.Logger, tracer opentracing.Tracer) {
	logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "Service"), zap.Any("context", "Start subscribe NATS..."))
	subj, i := "*." + Cfg.ServiceName + ".>", 0
	// Подписываемся
	subq, _ := Ec.QueueSubscribe(subj, Cfg.ServiceName + "-queue", func(subject, reply string, request *RequestMessage) {
		i++
		logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "Service"), zap.Any("context", "[#" + fmt.Sprint(i) +"] Received on [" + subject + "]"), zap.Reflect("request", request))
		conn, err := Db_sub.Acquire(ctx)
		if err != nil {
			logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "Service"), zap.Any("context", "Error acquiring connection"), zap.Error(err))
		}
		defer conn.Release()
		var res, mq_id int
		// сохраним реквизиты полученного сообщения
		mq_id = int(request.Headers["mq_id"].(float64))
		err_query := conn.QueryRow(ctx, "insert into " + Cfg.ServiceName + "_service.mq_receive(mq_id,service_name) values ($1,$2) returning mq_id", mq_id, request.Headers["from_service"]).Scan(&res)
		// TODO, здесь при ошибке "ERROR: duplicate key value violates unique constraint \"fk_analytic_mq_receive\" (SQLSTATE 23505)" надо пропускать ход. Оставил для тестов массовых запросов
		if err_query != nil {
			logger.Warn("", zap.String("service", Cfg.ServiceName), zap.String("function", "Service"), zap.Any("context", "Error query"), zap.Int("reply_mq_id", mq_id), zap.Error(err_query))
		}
		if _, ok := MapHandlersFuncSubs[subject]; ok {
			Wg_sub.Add(1)
			go func(){
				defer Wg_sub.Done()
				// ошибку возвращать и обрабатывать
				res, err := HubFuncS(request, MapHandlersFuncSubs[subject], logger)
				var response map[string]interface{}
				if err != nil {
					response = map[string]interface{}{
						"error":     err,
						"meta": request.Headers,
					}
				} else {
					response = map[string]interface{}{
						"result":     res,
						"meta": request.Headers,
					}
				}
				logger.Debug("", zap.String("service", Cfg.ServiceName), zap.String("function", "Service"), zap.Any("context", "sending answer"), zap.Int("reply_mq_id", mq_id))
				Ec.Publish(reply, response)
			}()
		} else {
			var response map[string]interface{}
			response = map[string]interface{}{
				"error":     "wrong handler",
				"meta": request.Headers,
			}
			logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "Service"), zap.Any("context", "wrong handler"), zap.Int("reply_mq_id", mq_id))
			Ec.Publish(reply, response)
		}
	})
	Subq = subq
	Ec.Flush()
	if err := Ec.LastError(); err != nil {
		logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", "Service"), zap.Any("context", "Error reply"), zap.Error(Ec.LastError()))
	}
	logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "Service"), zap.Any("context", "Listening on subject [" + subj + "]"))
	listen(ctx, logger, tracer)
}
// Вспомогательная функция создания спанов
func createSpan (oper string, nameFunc string, header map[string]interface{}, logger *zap.Logger) opentracing.Span {
	var span opentracing.Span
	if x, found := header["trace-id"]; found {
		if trace_id, ok := x.(string); ok {
			carrier := opentracing.TextMapCarrier{}
			carrier.Set("uber-trace-id", trace_id)
			wireContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, carrier)
			if err != nil {
				span = opentracing.StartSpan(oper)
				logger.Error("", zap.String("service", Cfg.ServiceName), zap.String("function", nameFunc), zap.Any("context", "trace_id error"), zap.Error(err))
			} else {
				span = opentracing.StartSpan(oper, opentracing.ChildOf(wireContext))
			}
		} else {
			span = Tracer.StartSpan(oper)
		}
	} else {
		span = Tracer.StartSpan(oper)
	}
	return span
}
