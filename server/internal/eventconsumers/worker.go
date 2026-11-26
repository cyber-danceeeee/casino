package eventconsumers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"casino/internal/eventbus"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"gorm.io/gorm"
)

type WorkerOptions struct {
	WorkerID     string
	PollInterval time.Duration
	BatchSize    int
	MaxAttempts  uint
	LeaseTimeout time.Duration
}

type Worker struct {
	db      *gorm.DB
	options WorkerOptions
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type RabbitWorker struct {
	service       *eventbus.Service
	publisher     eventbus.EventPublisher
	consumer      eventbus.EventConsumer
	registrations []eventbus.ConsumerRegistration
	eventTypes    []string
	options       WorkerOptions
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

func NewRabbitWorker(db *gorm.DB, publisher eventbus.EventPublisher, consumer eventbus.EventConsumer, registrations []eventbus.ConsumerRegistration, options WorkerOptions) *RabbitWorker {
	if options.WorkerID == "" {
		options.WorkerID = fmt.Sprintf("rabbitmq-outbox-%d", time.Now().UnixNano())
	}
	if options.PollInterval <= 0 {
		options.PollInterval = time.Second
	}
	if options.BatchSize <= 0 || options.BatchSize > 500 {
		options.BatchSize = 100
	}
	if len(registrations) == 0 {
		registrations = Registrations(db)
	}
	types := make([]string, 0, len(registrations))
	seen := map[string]struct{}{}
	for _, registration := range registrations {
		for _, eventType := range registration.EventTypes {
			if _, ok := seen[eventType]; ok {
				continue
			}
			seen[eventType] = struct{}{}
			types = append(types, eventType)
		}
	}
	return &RabbitWorker{
		service: eventbus.NewWithOptions(db, eventbus.Options{
			WorkerID: options.WorkerID, MaxAttempts: options.MaxAttempts, LeaseTimeout: options.LeaseTimeout,
		}),
		publisher: publisher, consumer: consumer, registrations: registrations, eventTypes: types, options: options,
	}
}

func (w *RabbitWorker) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	if err := w.consumer.Start(ctx, w.registrations, w.service); err != nil {
		cancel()
		return err
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.options.PollInterval)
		defer ticker.Stop()
		w.processOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.processOnce(ctx)
			case <-eventbus.DispatchWake():
				w.processOnce(ctx)
			}
		}
	}()
	return nil
}

func (w *RabbitWorker) Stop() error {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	return w.consumer.Close()
}

func (w *RabbitWorker) processOnce(ctx context.Context) {
	summary, err := w.service.RelayPending(ctx, w.publisher, w.eventTypes, w.options.BatchSize)
	if err != nil {
		logger.Errorf("[rabbitmq-outbox-worker] relay events: %v", err)
		return
	}
	if summary.Published > 0 || summary.Failed > 0 || summary.Dead > 0 {
		logger.Infof("[rabbitmq-outbox-worker] relay summary: scanned=%d published=%d failed=%d dead=%d", summary.Scanned, summary.Published, summary.Failed, summary.Dead)
	}
}

func NewWorker(db *gorm.DB, options WorkerOptions) *Worker {
	if options.WorkerID == "" {
		options.WorkerID = fmt.Sprintf("event-consumer-%d", time.Now().UnixNano())
	}
	if options.PollInterval <= 0 {
		options.PollInterval = 5 * time.Second
	}
	if options.BatchSize <= 0 || options.BatchSize > 500 {
		options.BatchSize = 100
	}
	return &Worker{db: db, options: options}
}

func (w *Worker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.options.PollInterval)
		defer ticker.Stop()
		w.ProcessOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.ProcessOnce(ctx)
			}
		}
	}()
}

func (w *Worker) Stop() error {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	return nil
}

func (w *Worker) ProcessOnce(ctx context.Context) {
	service, err := NewService(w.db, eventbus.Options{
		WorkerID: w.options.WorkerID, MaxAttempts: w.options.MaxAttempts, LeaseTimeout: w.options.LeaseTimeout,
	})
	if err != nil {
		logger.Errorf("[event-consumer-worker] initialize consumers: %v", err)
		return
	}
	summary, err := service.DispatchPending(ctx, w.options.BatchSize)
	if err != nil {
		logger.Errorf("[event-consumer-worker] dispatch events: %v", err)
		return
	}
	if summary.Published > 0 || summary.Failed > 0 || summary.Dead > 0 {
		logger.Infof("[event-consumer-worker] dispatch summary: scanned=%d published=%d failed=%d dead=%d skipped=%d", summary.Scanned, summary.Published, summary.Failed, summary.Dead, summary.Skipped)
	}
}
