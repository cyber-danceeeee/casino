package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"casino/internal/eventbus"

	"github.com/rabbitmq/amqp091-go"
)

const persistentDeliveryMode = uint8(amqp091.Persistent)

var retrySteps = []struct {
	name  string
	delay time.Duration
}{
	{name: "5s", delay: 5 * time.Second},
	{name: "30s", delay: 30 * time.Second},
	{name: "2m", delay: 2 * time.Minute},
	{name: "10m", delay: 10 * time.Minute},
	{name: "30m", delay: 30 * time.Minute},
}

type Config struct {
	Host             string
	Port             int
	ManagementPort   int
	Username         string
	Password         string
	VHost            string
	TopicExchange    string
	DirectExchange   string
	FanoutExchange   string
	RetryExchange    string
	DeadExchange     string
	Prefetch         int
	Concurrency      int
	PublisherConfirm bool
	PublisherTimeout time.Duration
	ReconnectDelay   time.Duration
	MaxRetries       int
	Queues           []string
	Bindings         map[string][]string
}

type Adapter struct {
	config Config

	connectionMu sync.Mutex
	connection   *amqp091.Connection
	publishMu    sync.Mutex

	consumerCancel context.CancelFunc
	consumerWG     sync.WaitGroup
	consumerCount  atomic.Int64
	closed         atomic.Bool

	statsMu      sync.Mutex
	startedAt    time.Time
	published    uint64
	consumed     uint64
	latencyTotal time.Duration
}

func New(config Config) *Adapter {
	applyDefaults(&config)
	return &Adapter{config: config, startedAt: time.Now()}
}

func applyDefaults(config *Config) {
	if strings.TrimSpace(config.Host) == "" {
		config.Host = "127.0.0.1"
	}
	if config.Port <= 0 {
		config.Port = 5672
	}
	if config.ManagementPort <= 0 {
		config.ManagementPort = 15672
	}
	if strings.TrimSpace(config.VHost) == "" {
		config.VHost = "/"
	}
	if strings.TrimSpace(config.TopicExchange) == "" {
		config.TopicExchange = "event.topic"
	}
	if strings.TrimSpace(config.DirectExchange) == "" {
		config.DirectExchange = "event.direct"
	}
	if strings.TrimSpace(config.FanoutExchange) == "" {
		config.FanoutExchange = "event.fanout"
	}
	if strings.TrimSpace(config.RetryExchange) == "" {
		config.RetryExchange = "event.delay"
	}
	if strings.TrimSpace(config.DeadExchange) == "" {
		config.DeadExchange = "event.dead"
	}
	if config.Prefetch <= 0 {
		config.Prefetch = 20
	}
	if config.Concurrency <= 0 {
		config.Concurrency = 2
	}
	if config.PublisherTimeout <= 0 {
		config.PublisherTimeout = 5 * time.Second
	}
	if config.ReconnectDelay <= 0 {
		config.ReconnectDelay = 5 * time.Second
	}
	if config.MaxRetries <= 0 || config.MaxRetries > len(retrySteps) {
		config.MaxRetries = len(retrySteps)
	}
	config.Queues = uniqueStrings(append(config.Queues,
		eventbus.QueuePayment, eventbus.QueueProvider, eventbus.QueuePromotion,
		eventbus.QueueNotification, eventbus.QueueAudit, eventbus.QueueSystemJobs,
		eventbus.QueueRetry, eventbus.QueueDead,
	))
}

func (a *Adapter) Available(ctx context.Context) error {
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		return err
	}
	if connection.IsClosed() {
		return errors.New("rabbitmq connection is closed")
	}
	return nil
}

func (a *Adapter) Publish(ctx context.Context, message eventbus.PublishMessage) error {
	headers := cloneHeaders(message.Headers)
	headers[eventbus.HeaderOutboxID] = int64(message.OutboxID)
	headers["x-event-id"] = message.EventID
	headers["x-sent-at"] = time.Now().UTC().Format(time.RFC3339Nano)
	deliveryMode := message.DeliveryMode
	if deliveryMode == 0 {
		deliveryMode = persistentDeliveryMode
	}
	contentType := strings.TrimSpace(message.ContentType)
	if contentType == "" {
		contentType = "application/json"
	}
	publishing := amqp091.Publishing{
		Headers: toTable(headers), ContentType: contentType, DeliveryMode: deliveryMode,
		MessageId: message.EventID, CorrelationId: message.CorrelationID, Type: message.MessageType,
		Timestamp: message.OccurredAt, Body: message.Body,
	}
	if err := a.publishConfirmed(ctx, a.config.TopicExchange, message.RoutingKey, publishing); err != nil {
		return err
	}
	a.statsMu.Lock()
	a.published++
	a.statsMu.Unlock()
	return nil
}

func (a *Adapter) publishConfirmed(ctx context.Context, exchange string, routingKey string, publishing amqp091.Publishing) error {
	a.publishMu.Lock()
	defer a.publishMu.Unlock()
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		return err
	}
	channel, err := connection.Channel()
	if err != nil {
		a.invalidateConnection(connection)
		return err
	}
	defer channel.Close()
	if a.config.PublisherConfirm {
		if err := channel.Confirm(false); err != nil {
			return err
		}
	}
	returns := channel.NotifyReturn(make(chan amqp091.Return, 1))
	confirms := channel.NotifyPublish(make(chan amqp091.Confirmation, 1))
	if err := channel.PublishWithContext(ctx, exchange, routingKey, true, false, publishing); err != nil {
		return err
	}
	if !a.config.PublisherConfirm {
		select {
		case returned := <-returns:
			return fmt.Errorf("rabbitmq message returned: %s", returned.ReplyText)
		default:
			return nil
		}
	}
	timer := time.NewTimer(a.config.PublisherTimeout)
	defer timer.Stop()
	for {
		select {
		case returned := <-returns:
			return fmt.Errorf("rabbitmq message returned: %s", returned.ReplyText)
		case confirmation := <-confirms:
			if !confirmation.Ack {
				return errors.New("rabbitmq publisher confirm was nack")
			}
			return nil
		case <-timer.C:
			return errors.New("rabbitmq publisher confirm timeout")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (a *Adapter) EnsureTopology(ctx context.Context) error {
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		return err
	}
	channel, err := connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	for _, exchange := range []struct{ name, kind string }{
		{a.config.DirectExchange, amqp091.ExchangeDirect},
		{a.config.TopicExchange, amqp091.ExchangeTopic},
		{a.config.FanoutExchange, amqp091.ExchangeFanout},
		{a.config.RetryExchange, amqp091.ExchangeTopic},
		{a.config.DeadExchange, amqp091.ExchangeTopic},
	} {
		if err := channel.ExchangeDeclare(exchange.name, exchange.kind, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare exchange %s: %w", exchange.name, err)
		}
	}
	for _, queue := range a.config.Queues {
		if queue == eventbus.QueueRetry || queue == eventbus.QueueDead {
			if _, err := channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
				return fmt.Errorf("declare queue %s: %w", queue, err)
			}
			continue
		}
		if err := a.declareBusinessQueue(channel, queue); err != nil {
			return err
		}
	}
	for queue, routingKeys := range a.config.Bindings {
		for _, routingKey := range uniqueStrings(routingKeys) {
			if err := channel.QueueBind(queue, routingKey, a.config.TopicExchange, false, nil); err != nil {
				return fmt.Errorf("bind queue %s to %s: %w", queue, routingKey, err)
			}
		}
	}
	return nil
}

func (a *Adapter) declareBusinessQueue(channel *amqp091.Channel, queue string) error {
	deadRoutingKey := queue + ".dead"
	arguments := amqp091.Table{
		"x-dead-letter-exchange":    a.config.DeadExchange,
		"x-dead-letter-routing-key": deadRoutingKey,
	}
	if _, err := channel.QueueDeclare(queue, true, false, false, false, arguments); err != nil {
		return fmt.Errorf("declare queue %s: %w", queue, err)
	}
	if err := channel.QueueBind(queue, queue, a.config.DirectExchange, false, nil); err != nil {
		return fmt.Errorf("bind direct queue %s: %w", queue, err)
	}
	deadQueue := queue + ".dead"
	if _, err := channel.QueueDeclare(deadQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dead queue %s: %w", deadQueue, err)
	}
	if err := channel.QueueBind(deadQueue, deadRoutingKey, a.config.DeadExchange, false, nil); err != nil {
		return fmt.Errorf("bind dead queue %s: %w", deadQueue, err)
	}
	for _, step := range retrySteps[:a.config.MaxRetries] {
		retryQueue := retryQueueName(queue, step.name)
		retryRoutingKey := retryRoutingKey(queue, step.name)
		retryArguments := amqp091.Table{
			"x-message-ttl":             int32(step.delay / time.Millisecond),
			"x-dead-letter-exchange":    a.config.DirectExchange,
			"x-dead-letter-routing-key": queue,
		}
		if _, err := channel.QueueDeclare(retryQueue, true, false, false, false, retryArguments); err != nil {
			return fmt.Errorf("declare retry queue %s: %w", retryQueue, err)
		}
		if err := channel.QueueBind(retryQueue, retryRoutingKey, a.config.RetryExchange, false, nil); err != nil {
			return fmt.Errorf("bind retry queue %s: %w", retryQueue, err)
		}
	}
	return nil
}

func (a *Adapter) Start(ctx context.Context, registrations []eventbus.ConsumerRegistration, dispatcher eventbus.EventDispatcher) error {
	if dispatcher == nil {
		return errors.New("event dispatcher is required")
	}
	if a.consumerCancel != nil {
		return errors.New("rabbitmq consumers already started")
	}
	if err := a.EnsureTopology(ctx); err != nil {
		return err
	}
	groups := map[string][]eventbus.ConsumerRegistration{}
	for _, registration := range registrations {
		if strings.TrimSpace(registration.Name) == "" || strings.TrimSpace(registration.Queue) == "" || registration.Handler == nil {
			return errors.New("invalid RabbitMQ consumer registration")
		}
		groups[registration.Queue] = append(groups[registration.Queue], registration)
	}
	consumerCtx, cancel := context.WithCancel(ctx)
	a.consumerCancel = cancel
	for queue, queueRegistrations := range groups {
		concurrency := a.config.Concurrency
		for _, registration := range queueRegistrations {
			if registration.Concurrency > concurrency {
				concurrency = registration.Concurrency
			}
		}
		for worker := 0; worker < concurrency; worker++ {
			a.consumerWG.Add(1)
			go a.consumeLoop(consumerCtx, queue, worker, queueRegistrations, dispatcher)
		}
	}
	return nil
}

func (a *Adapter) consumeLoop(ctx context.Context, queue string, worker int, registrations []eventbus.ConsumerRegistration, dispatcher eventbus.EventDispatcher) {
	defer a.consumerWG.Done()
	for ctx.Err() == nil && !a.closed.Load() {
		if err := a.consumeChannel(ctx, queue, worker, registrations, dispatcher); err != nil && ctx.Err() == nil {
			if !sleepContext(ctx, a.config.ReconnectDelay) {
				return
			}
		}
	}
}

func (a *Adapter) consumeChannel(ctx context.Context, queue string, worker int, registrations []eventbus.ConsumerRegistration, dispatcher eventbus.EventDispatcher) error {
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		return err
	}
	channel, err := connection.Channel()
	if err != nil {
		a.invalidateConnection(connection)
		return err
	}
	defer channel.Close()
	if err := channel.Qos(a.config.Prefetch, 0, false); err != nil {
		return err
	}
	tag := fmt.Sprintf("casino-%s-%d-%d", strings.ReplaceAll(queue, ".", "-"), worker, time.Now().UnixNano())
	deliveries, err := channel.Consume(queue, tag, false, false, false, false, nil)
	if err != nil {
		return err
	}
	a.consumerCount.Add(1)
	defer a.consumerCount.Add(-1)
	closed := channel.NotifyClose(make(chan *amqp091.Error, 1))
	for {
		select {
		case <-ctx.Done():
			_ = channel.Cancel(tag, false)
			return nil
		case closeErr := <-closed:
			if closeErr == nil {
				return errors.New("rabbitmq consumer channel closed")
			}
			return closeErr
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("rabbitmq delivery channel closed")
			}
			a.handleDelivery(ctx, queue, registrations, dispatcher, delivery)
		}
	}
}

func (a *Adapter) handleDelivery(ctx context.Context, queue string, registrations []eventbus.ConsumerRegistration, dispatcher eventbus.EventDispatcher, delivery amqp091.Delivery) {
	message := eventbus.DeliveryMessage{
		Queue: queue, RoutingKey: delivery.RoutingKey, Body: delivery.Body,
		Headers: fromTable(delivery.Headers), ReceivedAt: time.Now(),
	}
	var dispatchErr error
	for _, registration := range registrations {
		if err := dispatcher.Dispatch(ctx, message, registration); err != nil {
			dispatchErr = fmt.Errorf("%s: %w", registration.Name, err)
			break
		}
	}
	if dispatchErr == nil {
		if err := delivery.Ack(false); err == nil {
			a.recordConsumed(delivery.Timestamp)
		}
		return
	}
	retryCount := headerInt(delivery.Headers, eventbus.HeaderRetryCount)
	if retryCount >= a.config.MaxRetries {
		_ = delivery.Reject(false)
		return
	}
	headers := fromTable(delivery.Headers)
	headers[eventbus.HeaderRetryCount] = int32(retryCount + 1)
	headers[eventbus.HeaderOriginalQueue] = queue
	headers[eventbus.HeaderLastError] = truncate(dispatchErr.Error(), 1000)
	step := retrySteps[retryCount]
	publishing := amqp091.Publishing{
		Headers: toTable(headers), ContentType: delivery.ContentType, ContentEncoding: delivery.ContentEncoding,
		DeliveryMode: persistentDeliveryMode, Priority: delivery.Priority, CorrelationId: delivery.CorrelationId,
		ReplyTo: delivery.ReplyTo, Expiration: "", MessageId: delivery.MessageId, Timestamp: delivery.Timestamp,
		Type: delivery.Type, UserId: delivery.UserId, AppId: delivery.AppId, Body: delivery.Body,
	}
	if err := a.publishConfirmed(ctx, a.config.RetryExchange, retryRoutingKey(queue, step.name), publishing); err != nil {
		_ = delivery.Nack(false, true)
		return
	}
	_ = delivery.Ack(false)
}

func (a *Adapter) recordConsumed(occurredAt time.Time) {
	now := time.Now()
	a.statsMu.Lock()
	a.consumed++
	if !occurredAt.IsZero() && occurredAt.Before(now) {
		a.latencyTotal += now.Sub(occurredAt)
	}
	a.statsMu.Unlock()
}

func (a *Adapter) Health(ctx context.Context) eventbus.QueueHealth {
	health := eventbus.QueueHealth{Status: "DOWN", ConsumerRunning: a.consumerCount.Load() > 0}
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		health.Message = err.Error()
		return health
	}
	health.Connection = !connection.IsClosed()
	health.PublisherAvailable = health.Connection
	channel, err := connection.Channel()
	if err != nil {
		health.Message = err.Error()
		return health
	}
	defer channel.Close()
	for _, exchange := range []struct{ name, kind string }{
		{a.config.TopicExchange, amqp091.ExchangeTopic}, {a.config.DirectExchange, amqp091.ExchangeDirect},
		{a.config.FanoutExchange, amqp091.ExchangeFanout}, {a.config.RetryExchange, amqp091.ExchangeTopic},
		{a.config.DeadExchange, amqp091.ExchangeTopic},
	} {
		if err := channel.ExchangeDeclarePassive(exchange.name, exchange.kind, true, false, false, false, nil); err != nil {
			health.MissingExchanges = append(health.MissingExchanges, exchange.name)
			break
		}
	}
	if len(health.MissingExchanges) == 0 {
		for _, queue := range a.config.Queues {
			if _, err := channel.QueueInspect(queue); err != nil {
				health.MissingQueues = append(health.MissingQueues, queue)
				break
			}
		}
	}
	if health.Connection && health.PublisherAvailable && len(health.MissingExchanges) == 0 && len(health.MissingQueues) == 0 {
		health.Status = "UP"
		health.Message = "RabbitMQ topology is available"
	} else if health.Connection {
		health.Status = "DEGRADED"
	}
	return health
}

func (a *Adapter) Metrics(ctx context.Context) (eventbus.QueueMetrics, error) {
	metrics, err := a.managementMetrics(ctx)
	if err != nil {
		metrics, err = a.inspectMetrics(ctx)
		if err != nil {
			return eventbus.QueueMetrics{}, err
		}
	}
	a.statsMu.Lock()
	elapsed := time.Since(a.startedAt).Seconds()
	publishTPS, consumeTPS, averageLatency := 0.0, 0.0, 0.0
	if elapsed > 0 {
		publishTPS = float64(a.published) / elapsed
		consumeTPS = float64(a.consumed) / elapsed
	}
	if a.consumed > 0 {
		averageLatency = float64(a.latencyTotal.Microseconds()) / 1000 / float64(a.consumed)
	}
	a.statsMu.Unlock()
	for index := range metrics.Queues {
		metrics.Queues[index].PublishTPS = publishTPS
		metrics.Queues[index].ConsumeTPS = consumeTPS
		metrics.Queues[index].AvgLatency = averageLatency
	}
	metrics.CapturedAt = time.Now()
	return metrics, nil
}

type managementQueue struct {
	Name                   string `json:"name"`
	Messages               int    `json:"messages"`
	MessagesReady          int    `json:"messages_ready"`
	MessagesUnacknowledged int    `json:"messages_unacknowledged"`
	Consumers              int    `json:"consumers"`
	MessageStats           struct {
		PublishDetails struct {
			Rate float64 `json:"rate"`
		} `json:"publish_details"`
		DeliverDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_details"`
	} `json:"message_stats"`
}

func (a *Adapter) managementMetrics(ctx context.Context) (eventbus.QueueMetrics, error) {
	endpoint := fmt.Sprintf("http://%s:%d/api/queues/%s", a.config.Host, a.config.ManagementPort, url.PathEscape(a.config.VHost))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return eventbus.QueueMetrics{}, err
	}
	request.SetBasicAuth(a.config.Username, a.config.Password)
	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return eventbus.QueueMetrics{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return eventbus.QueueMetrics{}, fmt.Errorf("rabbitmq management returned %s", response.Status)
	}
	rows := []managementQueue{}
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return eventbus.QueueMetrics{}, err
	}
	result := eventbus.QueueMetrics{Queues: []eventbus.QueueMetric{}}
	for _, row := range rows {
		if !a.managedQueue(row.Name) {
			continue
		}
		metric := eventbus.QueueMetric{
			Queue: row.Name, Depth: row.Messages, Ready: row.MessagesReady,
			Unacked: row.MessagesUnacknowledged, Consumers: row.Consumers,
			PublishTPS: row.MessageStats.PublishDetails.Rate, ConsumeTPS: row.MessageStats.DeliverDetails.Rate,
		}
		result.Queues = append(result.Queues, metric)
		if strings.HasSuffix(row.Name, ".dead") || row.Name == eventbus.QueueDead {
			result.DeadMessages += row.Messages
		}
		if strings.Contains(row.Name, ".retry.") || row.Name == eventbus.QueueRetry {
			result.RetryMessages += row.Messages
		}
	}
	sort.Slice(result.Queues, func(i, j int) bool { return result.Queues[i].Queue < result.Queues[j].Queue })
	return result, nil
}

func (a *Adapter) inspectMetrics(ctx context.Context) (eventbus.QueueMetrics, error) {
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		return eventbus.QueueMetrics{}, err
	}
	channel, err := connection.Channel()
	if err != nil {
		return eventbus.QueueMetrics{}, err
	}
	defer channel.Close()
	result := eventbus.QueueMetrics{Queues: []eventbus.QueueMetric{}}
	for _, queue := range a.allQueueNames() {
		inspected, inspectErr := channel.QueueInspect(queue)
		if inspectErr != nil {
			continue
		}
		metric := eventbus.QueueMetric{Queue: queue, Depth: inspected.Messages, Ready: inspected.Messages, Consumers: inspected.Consumers}
		result.Queues = append(result.Queues, metric)
		if strings.HasSuffix(queue, ".dead") || queue == eventbus.QueueDead {
			result.DeadMessages += inspected.Messages
		}
		if strings.Contains(queue, ".retry.") || queue == eventbus.QueueRetry {
			result.RetryMessages += inspected.Messages
		}
	}
	return result, nil
}

func (a *Adapter) DeadMessages(ctx context.Context, queue string, limit int) ([]eventbus.DeadMessage, error) {
	queue = normalizeDeadQueue(queue)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		return nil, err
	}
	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	defer channel.Close()
	deliveries := make([]amqp091.Delivery, 0, limit)
	result := make([]eventbus.DeadMessage, 0, limit)
	for index := 0; index < limit; index++ {
		delivery, ok, getErr := channel.Get(queue, false)
		if getErr != nil {
			return nil, getErr
		}
		if !ok {
			break
		}
		deliveries = append(deliveries, delivery)
		result = append(result, deadMessage(queue, delivery))
	}
	for _, delivery := range deliveries {
		_ = delivery.Nack(false, true)
	}
	return result, nil
}

func (a *Adapter) ReplayDead(ctx context.Context, queue string, eventID string) error {
	deadQueue := normalizeDeadQueue(queue)
	mainQueue := strings.TrimSuffix(deadQueue, ".dead")
	if mainQueue == deadQueue || !contains(a.config.Queues, mainQueue) {
		return errors.New("invalid dead-letter queue")
	}
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		return err
	}
	channel, err := connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	held := []amqp091.Delivery{}
	requeueHeld := func() {
		for _, item := range held {
			_ = item.Nack(false, true)
		}
	}
	for index := 0; index < 500; index++ {
		delivery, ok, getErr := channel.Get(deadQueue, false)
		if getErr != nil {
			return getErr
		}
		if !ok {
			break
		}
		if eventID != "" && delivery.MessageId != eventID {
			held = append(held, delivery)
			continue
		}
		headers := fromTable(delivery.Headers)
		headers[eventbus.HeaderRetryCount] = int32(0)
		delete(headers, eventbus.HeaderLastError)
		publishing := amqp091.Publishing{
			Headers: toTable(headers), ContentType: delivery.ContentType, DeliveryMode: persistentDeliveryMode,
			CorrelationId: delivery.CorrelationId, MessageId: delivery.MessageId, Timestamp: delivery.Timestamp,
			Type: delivery.Type, Body: delivery.Body,
		}
		if err := a.publishConfirmed(ctx, a.config.DirectExchange, mainQueue, publishing); err != nil {
			_ = delivery.Nack(false, true)
			requeueHeld()
			return err
		}
		if err := delivery.Ack(false); err != nil {
			requeueHeld()
			return err
		}
		requeueHeld()
		return nil
	}
	requeueHeld()
	return eventbus.ErrEventNotFound
}

func (a *Adapter) PurgeRetry(ctx context.Context, queue string) (int, error) {
	queue = strings.TrimSpace(queue)
	if !contains(a.config.Queues, queue) || queue == eventbus.QueueRetry || queue == eventbus.QueueDead {
		return 0, errors.New("invalid retry queue")
	}
	connection, err := a.ensureConnection(ctx)
	if err != nil {
		return 0, err
	}
	channel, err := connection.Channel()
	if err != nil {
		return 0, err
	}
	defer channel.Close()
	total := 0
	for _, step := range retrySteps[:a.config.MaxRetries] {
		count, purgeErr := channel.QueuePurge(retryQueueName(queue, step.name), false)
		if purgeErr != nil {
			return total, purgeErr
		}
		total += count
	}
	return total, nil
}

func (a *Adapter) Close() error {
	if !a.closed.CompareAndSwap(false, true) {
		return nil
	}
	if a.consumerCancel != nil {
		a.consumerCancel()
	}
	a.consumerWG.Wait()
	a.connectionMu.Lock()
	defer a.connectionMu.Unlock()
	if a.connection != nil && !a.connection.IsClosed() {
		return a.connection.Close()
	}
	return nil
}

func (a *Adapter) ensureConnection(ctx context.Context) (*amqp091.Connection, error) {
	if a.closed.Load() {
		return nil, errors.New("rabbitmq adapter is closed")
	}
	a.connectionMu.Lock()
	defer a.connectionMu.Unlock()
	if a.connection != nil && !a.connection.IsClosed() {
		return a.connection, nil
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	connection, err := amqp091.DialConfig(a.connectionURL(), amqp091.Config{
		Heartbeat: 10 * time.Second, Locale: "en_US",
		Dial: func(network, address string) (net.Conn, error) { return dialer.DialContext(ctx, network, address) },
	})
	if err != nil {
		return nil, err
	}
	a.connection = connection
	return connection, nil
}

func (a *Adapter) invalidateConnection(connection *amqp091.Connection) {
	a.connectionMu.Lock()
	defer a.connectionMu.Unlock()
	if a.connection == connection {
		_ = a.connection.Close()
		a.connection = nil
	}
}

func (a *Adapter) connectionURL() string {
	path := a.config.VHost
	if path != "/" {
		path = "/" + strings.TrimPrefix(path, "/")
	}
	value := &url.URL{
		Scheme: "amqp", Host: fmt.Sprintf("%s:%d", a.config.Host, a.config.Port), Path: path,
		User: url.UserPassword(a.config.Username, a.config.Password),
	}
	return value.String()
}

func (a *Adapter) allQueueNames() []string {
	values := append([]string{}, a.config.Queues...)
	for _, queue := range a.config.Queues {
		if queue == eventbus.QueueRetry || queue == eventbus.QueueDead {
			continue
		}
		values = append(values, queue+".dead")
		for _, step := range retrySteps[:a.config.MaxRetries] {
			values = append(values, retryQueueName(queue, step.name))
		}
	}
	return uniqueStrings(values)
}

func (a *Adapter) managedQueue(queue string) bool {
	return contains(a.allQueueNames(), queue)
}

func retryQueueName(queue string, step string) string  { return queue + ".retry." + step }
func retryRoutingKey(queue string, step string) string { return queue + ".retry." + step }

func normalizeDeadQueue(queue string) string {
	queue = strings.TrimSpace(queue)
	if !strings.HasSuffix(queue, ".dead") {
		queue += ".dead"
	}
	return queue
}

func deadMessage(queue string, delivery amqp091.Delivery) eventbus.DeadMessage {
	payload := map[string]any{}
	_ = json.Unmarshal(delivery.Body, &payload)
	eventType := delivery.Type
	if eventType == "" {
		eventType, _ = payload["event_type"].(string)
	}
	return eventbus.DeadMessage{
		Queue: queue, EventID: delivery.MessageId, EventType: eventType, RoutingKey: delivery.RoutingKey,
		RetryCount: headerInt(delivery.Headers, eventbus.HeaderRetryCount),
		Error:      fmt.Sprint(delivery.Headers[eventbus.HeaderLastError]), Payload: payload, CreatedAt: delivery.Timestamp,
	}
}

func toTable(headers map[string]any) amqp091.Table {
	table := amqp091.Table{}
	for key, value := range headers {
		switch typed := value.(type) {
		case uint64:
			table[key] = int64(typed)
		case uint:
			table[key] = int64(typed)
		case int:
			table[key] = int64(typed)
		default:
			table[key] = typed
		}
	}
	return table
}

func fromTable(table amqp091.Table) map[string]any {
	values := map[string]any{}
	for key, value := range table {
		values[key] = value
	}
	return values
}

func cloneHeaders(headers map[string]any) map[string]any {
	values := map[string]any{}
	for key, value := range headers {
		values[key] = value
	}
	return values
}

func headerInt(headers amqp091.Table, key string) int {
	value := headers[key]
	switch typed := value.(type) {
	case int:
		return typed
	case int8:
		return int(typed)
	case int16:
		return int(typed)
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(typed)
		return parsed
	}
	return 0
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func truncate(value string, length int) string {
	if len(value) <= length {
		return value
	}
	return value[:length]
}

func sleepContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
