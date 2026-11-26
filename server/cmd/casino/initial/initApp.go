// Package initial is the package that starts the service to initialize the service, including
// the initialization configuration, service configuration, connecting to the database, and
// resource release needed when shutting down the service.
package initial

import (
	"context"
	"flag"
	"strconv"
	"time"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/stat"
	"github.com/go-dev-frame/sponge/pkg/tracer"

	"casino/configs"
	"casino/internal/config"
	"casino/internal/database"
	"casino/internal/eventbus"
	rabbitadapter "casino/internal/eventbus/rabbitmq"
	"casino/internal/eventconsumers"
	"casino/internal/notificationpush"
	"casino/internal/systemadmin"
)

var (
	version      string
	configFile   string
	pushWorker   *notificationpush.Worker
	eventWorker  *eventconsumers.Worker
	rabbitWorker *eventconsumers.RabbitWorker
)

// InitApp initial app configuration
func InitApp() {
	initConfig()
	cfg := config.Get()

	// initializing log
	_, err := logger.Init(
		logger.WithLevel(cfg.Logger.Level),
		logger.WithFormat(cfg.Logger.Format),
		logger.WithSave(
			cfg.Logger.IsSave,
			//logger.WithFileName(cfg.Logger.LogFileConfig.Filename),
			//logger.WithFileMaxSize(cfg.Logger.LogFileConfig.MaxSize),
			//logger.WithFileMaxBackups(cfg.Logger.LogFileConfig.MaxBackups),
			//logger.WithFileMaxAge(cfg.Logger.LogFileConfig.MaxAge),
			//logger.WithFileIsCompression(cfg.Logger.LogFileConfig.IsCompression),
		),
	)
	if err != nil {
		panic(err)
	}
	logger.Debug(config.Show())
	logger.Info("[logger] was initialized")

	// initializing tracing
	if cfg.App.EnableTrace {
		tracer.InitWithConfig(
			cfg.App.Name,
			cfg.App.Env,
			cfg.App.Version,
			cfg.Jaeger.AgentHost,
			strconv.Itoa(cfg.Jaeger.AgentPort),
			cfg.App.TracingSamplingRate,
		)
		logger.Info("[tracer] was initialized")
	}

	// initializing the print system and process resources
	if cfg.App.EnableStat {
		stat.Init(
			stat.WithLog(logger.Get()),
			stat.WithAlarm(), // invalid if it is windows, the default threshold for cpu and memory is 0.8, you can modify them
			stat.WithPrintField(logger.String("service_name", cfg.App.Name), logger.String("host", cfg.App.Host)),
		)
		logger.Info("[resource statistics] was initialized")
	}

	// initializing database
	database.InitDB()
	if err := systemadmin.New(database.GetDB()).EnsureSiteConfig(); err != nil {
		panic(err)
	}
	logger.Infof("[%s] was initialized", cfg.Database.Driver)
	//if err := notificationpush.EnsureSchema(database.GetDB()); err != nil {
	//	panic("init push notification schema error: " + err.Error())
	//}
	//if err := eventconsumers.EnsureSchema(database.GetDB()); err != nil {
	//	panic("init event consumer schema error: " + err.Error())
	//}
	pushWorker = notificationpush.New(database.GetDB(), cfg.Notification.FCM)
	pushWorker.Start()
	if err := pushWorker.InitializationError(); err != nil {
		logger.Errorf("[notification worker] Firebase initialization failed: %v", err)
	} else {
		logger.Info("[notification worker] Firebase Messaging was initialized")
	}
	initializeEventBus(cfg)
	database.InitCache(cfg.App.CacheType)
	if cfg.App.CacheType != "" {
		logger.Infof("[%s] was initialized", cfg.App.CacheType)
	}
}

func initializeEventBus(cfg *config.Config) {
	rabbitConfig := config.RabbitMQConfig()
	if !rabbitConfig.Enabled {
		logger.Warn("[event bus] RabbitMQ is disabled; outbox delivery is paused")
		return
	}
	registrations := eventconsumers.Registrations(database.GetDB(), pushWorker)
	bindings := map[string][]string{}
	queues := []string{}
	for _, registration := range registrations {
		bindings[registration.Queue] = append(bindings[registration.Queue], registration.RoutingKeys...)
		queues = append(queues, registration.Queue)
	}
	adapter := rabbitadapter.New(rabbitadapter.Config{
		Host: rabbitConfig.Host, Port: rabbitConfig.Port, ManagementPort: rabbitConfig.ManagementPort,
		Username: rabbitConfig.Username, Password: rabbitConfig.Password, VHost: rabbitConfig.VHost,
		TopicExchange: rabbitConfig.Exchange, DirectExchange: rabbitConfig.DirectExchange,
		FanoutExchange: rabbitConfig.FanoutExchange, RetryExchange: rabbitConfig.RetryExchange, DeadExchange: rabbitConfig.DeadExchange,
		Prefetch: rabbitConfig.Prefetch, Concurrency: rabbitConfig.Concurrency, PublisherConfirm: rabbitConfig.PublisherConfirm,
		PublisherTimeout: time.Duration(rabbitConfig.PublisherTimeoutMS) * time.Millisecond,
		ReconnectDelay:   time.Duration(rabbitConfig.ReconnectDelaySeconds) * time.Second,
		MaxRetries:       rabbitConfig.MaxRetries, Queues: queues, Bindings: bindings,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := adapter.EnsureTopology(ctx); err != nil {
		_ = adapter.Close()
		if rabbitConfig.Required {
			panic("init RabbitMQ topology error: " + err.Error())
		}
		logger.Errorf("[event bus] RabbitMQ initialization failed: %v", err)
		return
	}
	eventbus.SetQueueManager(adapter)
	rabbitWorker = eventconsumers.NewRabbitWorker(database.GetDB(), adapter, adapter, registrations, eventconsumers.WorkerOptions{})
	if err := rabbitWorker.Start(); err != nil {
		_ = adapter.Close()
		if rabbitConfig.Required {
			panic("start RabbitMQ worker error: " + err.Error())
		}
		logger.Errorf("[event bus] RabbitMQ worker failed: %v", err)
		return
	}
	logger.Info("[event bus] RabbitMQ adapter, outbox relay, and consumers were initialized")
}

func initConfig() {
	flag.StringVar(&version, "version", "", "service Version Number")
	flag.StringVar(&configFile, "c", "", "configuration file")
	flag.Parse()

	getConfigFromLocal()

	if version != "" {
		config.Get().App.Version = version
	}
}

// get configuration from local configuration file
func getConfigFromLocal() {
	if configFile == "" {
		configFile = configs.Location("casino.yml")
	}
	err := config.Init(configFile)
	if err != nil {
		panic("init config error: " + err.Error())
	}
}
