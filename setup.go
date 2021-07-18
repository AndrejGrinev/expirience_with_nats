package lib

import (
	"github.com/jessevdk/go-flags"
	"strconv"
	"github.com/jackc/pgx/v4/pgxpool"
	"context"
	"github.com/nats-io/nats.go"
	"time"
	"errors"
	"sync"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"fmt"
	"io"
	opentracing "github.com/opentracing/opentracing-go"
	config "github.com/uber/jaeger-client-go/config"
//	logzap "github.com/uber/jaeger-client-go/log/zap"
	"os"
	"os/signal"
	"syscall"
)

var (
	// ErrGotHelp returned after showing requested help
	ErrGotHelp = errors.New("help printed")
	// ErrBadArgs returned after showing command args error message
	ErrBadArgs = errors.New("option error printed")
)

var Cfg *Config

var Db_req, Db_sub, Db_sht *pgxpool.Pool

var Nc *nats.Conn

var Wg_req, Wg_sub sync.WaitGroup

var Ec *nats.EncodedConn

var Subq *nats.Subscription

var Tracer opentracing.Tracer

var Closer io.Closer

var Stopped bool = false

var CountNotify int = 0
// Конфиг
type Config struct {
	ServiceVersion string  `long:"service_version"  json:"service_version"  default:"1.01"                       description:"версия сервиса"`
	PoolMax        string  `long:"pool_max"         json:"pool_max"         default:"0"                          description:"Макс. коннектов к базе в одном пуле, по умолчание равно NUMCPU"`
	ServiceName    string  `long:"service_name"     json:"service_name"     default:"main"                       description:"Название сервиса"`
	UrlNATS        string  `long:"url_nats"         json:"url_nats"         default:"nats_auth.iac.tender.pro"   description:"урл nats"`
	TokenNATS      string  `long:"token_nats"       json:"token_nats"       default:"dag0HTXl4RGg7dXdaJwbC8"     description:"Bcrypted Tokens nats"`
	PgUser         string  `long:"pguser"           json:"pguser"           default:"usetender"                  description:"Postgres user"`
	PgHost         string  `long:"pghost"           json:"pghost"           default:"localhost"                  description:"Postgres host"`
	PgPort         string  `long:"pgport"           json:"pgport"           default:"5499"                       description:"Postgres port"`
	PgPassword     string  `long:"pgpassword"       json:"pgpassword"       default:""                           description:"Postgres password"`
	PgDbName       string  `long:"pgdbname"         json:"pgdbname"         default:"usetender"                  description:"Postgres database"`
	HostJaeger     string  `long:"host_jaeger"      json:"host_jaeger"      default:"jaeger"                     description:"Agent host jaeger"`
	PortJaeger     string  `long:"port_jaeger"      json:"port_jaeger"      default:"6831"                       description:"Agent port jaeger"`
	Debug          string  `long:"debug"            json:"debug"            default:"false"                      description:"Уровень вывода в лог, дебаг - true/false"`
	LogFileName    string  `long:"log_file_name"    json:"log_file_name"    default:"./main_service.log"         description:"Наименование и локация лога"`
	MaxSize        int     `long:"max_size"         json:"max_size"         default:"10"                         description:"Макс.размер лога(мб)"`
	MaxBackups     int     `long:"max_backups"      json:"max_backups"      default:"5"                          description:"Кол-во бэкапов"`
	MaxAge         int     `long:"max_age"          json:"max_age"          default:"30"                         description:"Макс.число дней хранения бэкапов"`
	Compress       string  `long:"compress"         json:"compress"         default:"false"                      description:"Бэкапы жмем? - true/false"`
	TypeJaeger     string  `long:"type_jaeger"      json:"type_jaeger"      default:"const"                      description:"todo"`
	ParamJaeger    float64 `long:"param_jaeger"     json:"param_jaeger"     default:"1"                          description:"todo"`
	LogSpansJaeger string  `long:"log_spans_jaeger" json:"log_spans_jaeger" default:"true"                       description:"todo"`
	TotalWaitNATS  int     `long:"total_wait_nats"  json:"total_wait_nats"  default:"10"                         description:"todo"`
	ReconDelayNATS int     `long:"recon_delay_nats" json:"recon_delay_nats" default:"1"                          description:"todo"`
}
// структура запроса
type RequestMessage struct {
	Headers map[string]interface{}
	Body    map[string]interface{}
}
// структура для канала ответа из транзита
type ChanMessage struct {
	Body    chan map[string]interface{}
	Transit chan bool
}
// setupConfig loads flags from args (if given) or command flags and ENV otherwise
func SetupConfig(args ...string) (*Config, error) {
	var err error

	cfg := &Config{}
	p := flags.NewParser(cfg, flags.Default) //  HelpFlag | PrintErrors | PassDoubleDash

	if len(args) == 0 {
		_, err = p.Parse()
	} else {
		_, err = p.ParseArgs(args)
	}
	if err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
			return nil, ErrGotHelp
		}
		return nil, err
	}

	return cfg, nil
}
// CONNECT TO DB
func SetupDb(logger *zap.Logger) (*pgxpool.Pool, error) {
	var strConfigDb string = ""
	// получаем пул соединений
	if Cfg.PoolMax != "0" {
		strConfigDb = "pool_max_conns=" + Cfg.PoolMax
	}
	config, _ := pgxpool.ParseConfig(strConfigDb)
	logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupDb"), zap.Any("DbPoolMaxConns", fmt.Sprint(config.MaxConns)))
	config.ConnConfig.Host = Cfg.PgHost
	n, err := strconv.ParseUint(Cfg.PgPort, 10, 16)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.Port = uint16(n)
	config.ConnConfig.User = Cfg.PgUser
	config.ConnConfig.Password = Cfg.PgPassword
	config.ConnConfig.Database = Cfg.PgDbName
	db, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
	return db, nil
}
// CONNECT TO NATS
func SetupNats(logger *zap.Logger) (*nats.Conn, error) {
	// Connect Options.
	opts := []nats.Option{nats.Name("NATS TPRO connect")}
	opts = setupConnOptions(opts, logger)
	// Connect to NATS
	nc, err := nats.Connect(Cfg.UrlNATS, opts...)
	if err != nil {
		return nil, err
	}
	return nc, nil
}
// Create NATS json encoded
func SetupNatsEncoded(nc *nats.Conn, logger *zap.Logger) (*nats.EncodedConn, error) {
	// Connect to NATS
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}
	return ec, nil
}
// setup options to nats connect
func setupConnOptions(opts []nats.Option, logger *zap.Logger) []nats.Option {
	totalWait := time.Duration(Cfg.TotalWaitNATS) * time.Minute
	reconnectDelay := time.Duration(Cfg.ReconDelayNATS) * time.Second

	opts = append(opts, nats.Token(Cfg.TokenNATS))
	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		logger.Warn("", zap.String("service", Cfg.ServiceName), zap.String("function", "setupConnOptions"), zap.Any("context", "Disconnected nats"), zap.Error(err), zap.Any("Disconnected, will attempt reconnects for minutes", totalWait.Minutes()))
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		logger.Warn("", zap.String("service", Cfg.ServiceName), zap.String("function", "setupConnOptions"), zap.Any("context", "Reconnected"), zap.Any("ConnectedUrl", nc.ConnectedUrl() ))
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "setupConnOptions"), zap.Any("context", "Exiting nats"), zap.Error(nc.LastError()))
	}))
	return opts
}
// setup log
func SetupLog(debug bool) *zap.Logger {

	encoder := getEncoder()
	writerSyncer := getLogWriter()

	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}
	core := zapcore.NewCore(encoder, writerSyncer, level)
	logger := zap.New(core, zap.AddCaller())
	return logger
}
// opts for encoder log
func getEncoder() zapcore.Encoder {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
}
// opts for writer log
func getLogWriter() zapcore.WriteSyncer {
	compress, _ := strconv.ParseBool(Cfg.Compress)
	lumberJackLogger := &lumberjack.Logger{
		Filename:   Cfg.LogFileName,
		MaxSize:    Cfg.MaxSize,
		MaxBackups: Cfg.MaxBackups,
		MaxAge:     Cfg.MaxAge,
		Compress:   compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}
// setup jaeger
func SetupJaeger(logger *zap.Logger) (tracer opentracing.Tracer, closer io.Closer, err error) {
	logspansjaeger, _ := strconv.ParseBool(Cfg.LogSpansJaeger)
	c := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  Cfg.TypeJaeger,
			Param: Cfg.ParamJaeger,
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: fmt.Sprintf("%s:%s", Cfg.HostJaeger, Cfg.PortJaeger),
			LogSpans:           logspansjaeger,
		},
	}
	tracer, closer, err = c.New(Cfg.ServiceName)
//	tracer, closer, err = c.New(Cfg.ServiceName, config.Logger(logzap.NewLogger(logger)))
	if err == nil {
		opentracing.SetGlobalTracer(tracer)
	}
	return tracer, closer, err
}
// Handler for graceful stop
func SetupCloseHandler(ctx context.Context, logger *zap.Logger, cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
	        sig := <-sigs
		Stopped = true
		logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupCloseHandler"), zap.Any("context", "Receive signal."), zap.Any("signal", sig))
		logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupCloseHandler"), zap.Any("context", "On start closed"), zap.Any("CountNotify", CountNotify))
		// дождемся когда отработают все нотифаи
		for {
			if CountNotify > 0 {
				logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupCloseHandler"), zap.Any("context", "In process closed"), zap.Any("CountNotify", CountNotify))
			} else {
				break
			}
			time.Sleep(time.Second)
		}
		logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupCloseHandler"), zap.Any("context", "Finish closed"), zap.Any("CountNotify", CountNotify))
		time.Sleep(time.Second)
		logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupCloseHandler"), zap.Any("context", "Send Cancel to ctx"))
		// send to context
		cancel()
		logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupCloseHandler"), zap.Any("context", "Draining nats..."))
		// drain nats messages
		if err := Subq.Drain(); err != nil {
			logger.Fatal("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupCloseHandler"), zap.Any("context", "DRAIN ERROR."), zap.Error(err))
		}
	}()
	logger.Info("", zap.String("service", Cfg.ServiceName), zap.String("function", "SetupCloseHandler"), zap.Any("context", "awaiting signal"))
}
