package main

import (
	"tpro/nats_service/lib"
	"time"
	"go.uber.org/zap"
	"log"
	"strconv"
	"github.com/satori/go.uuid"
	"fmt"
	jaeger "github.com/uber/jaeger-client-go"
	"context"
)

func main() {
	var err error
	// получим данные из конфига
	lib.Cfg, err = lib.SetupConfig()
	if err != nil {
		log.Fatalf( "Setup config: %v\n",err)
	}
	log.Printf("Get config:%+v\n", lib.Cfg)
	// подключаем лог
	debug, _ := strconv.ParseBool(lib.Cfg.Debug)
	logger := lib.SetupLog(debug)
	defer logger.Sync()
	logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.Any("context", "logger created."))
	// подключаем jaeger
	lib.Tracer, lib.Closer, err = lib.SetupJaeger(logger)
	if err != nil {
		logger.Fatal("", zap.String("context", "Setup jaeger"), zap.Error(err))
	}
	defer lib.Closer.Close()
	logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "Jaeger connected."))
	// jaeger health-check
	span := lib.Tracer.StartSpan("Start tpro nats microservice:"+ lib.Cfg.ServiceName)
	span.LogKV("started", "ya-ya!")
	sctx, ok := span.Context().(jaeger.SpanContext)
	if ok {
	        logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "jaeger health-check OK."), zap.String("test span", sctx.String()))
	} else {
		logger.Fatal("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "jaeger health-check ERROR."))
	}
	span.Finish()
	// законнектимся к серверу NATS
	for {
		lib.Nc, err = lib.SetupNats(logger)
		if err != nil {
			logger.Error("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "Setup nats"), zap.Error(err))
		} else {
			break
		}
		time.Sleep(time.Second)
	}
	defer lib.Nc.Close()
	logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "NATS connected."))
	// NATS health-check
	u := uuid.Must(uuid.NewV4(), err)
	subj := fmt.Sprintf("$$.nats-health-check.%s", u)
	sub, _ := lib.Nc.SubscribeSync(subj)
	_ = lib.Nc.Publish(subj, nil)
	_, err = sub.NextMsg(1e+10)
	if err != nil {
		logger.Fatal("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "NATS health-check ERROR."), zap.Error(err))
	}
	logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "NATS health-check OK."), zap.Any("test subj", u))
	// encode к коннекту nats 
	for {
		lib.Ec, err = lib.SetupNatsEncoded(lib.Nc, logger)
		if err != nil {
			logger.Error("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "Setup nats json encoded"), zap.Error(err))
		} else {
			break
		}
		time.Sleep(time.Second)
	}
	defer lib.Ec.Close()
	logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "NATS json encoded created."))
	// законнектимся к базе, получим пул коннектов для request-сообщений
	for {
		lib.Db_req, err = lib.SetupDb(logger)
		if err != nil {
			logger.Error("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "Setup db for request"), zap.Error(err))
		} else {
			break
		}
		time.Sleep(time.Second)
	}
	defer lib.Db_req.Close()
	logger.Info("", zap.Any("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "Получили пул коннектов DB для request."))
	// законнектимся к базе, получим пул коннектов для subscribe-сообщений
	for {
		lib.Db_sub, err = lib.SetupDb(logger)
		if err != nil {
			logger.Error("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "Setup db for subscribe"), zap.Error(err))
		} else {
			break
		}
		time.Sleep(time.Second)
	}
	defer lib.Db_sub.Close()
	logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "Получили пул коннектов DB для subscribe."))
	logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.String("context", "Start tpro nats microservice"), zap.String("version", lib.Cfg.ServiceVersion))
	// определяем контекст
	ctx, cancel := context.WithCancel(context.Background())
	// ловим сигналы на стоп сервиса
	lib.SetupCloseHandler(ctx, logger, cancel)
	// Запускаем сервис. Слушаем нотифаи (задания для request) & подписываемся на сообщения для нашего сервиса
	lib.Service(ctx, logger, lib.Tracer)
	lib.Wg_req.Wait()
	lib.Wg_sub.Wait()
	logger.Info("", zap.String("service", lib.Cfg.ServiceName), zap.String("function", "main"), zap.Any("context", "Shutting down!"))
}
