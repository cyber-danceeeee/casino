package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"casino/internal/eventbus"

	"github.com/rabbitmq/amqp091-go"
)

type integrationDispatcher struct {
	mu             sync.Mutex
	calls          map[string]int
	businessWrites map[string]int
	alwaysFail     map[string]bool
}

func (d *integrationDispatcher) Dispatch(_ context.Context, message eventbus.DeliveryMessage, _ eventbus.ConsumerRegistration) error {
	var envelope eventbus.EventEnvelope
	if err := json.Unmarshal(message.Body, &envelope); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls[message.Queue]++
	if d.alwaysFail[message.Queue] || (message.Queue == "integration.rabbit.retry" && d.calls[message.Queue] == 1) {
		return errors.New("integration retry")
	}
	if d.businessWrites[envelope.EventID] == 0 {
		d.businessWrites[envelope.EventID] = 1
	}
	return nil
}

func TestRabbitMQIntegrationRetryDeadReplayReconnectAndShutdown(t *testing.T) {
	if os.Getenv("RABBITMQ_INTEGRATION") != "1" {
		t.Skip("set RABBITMQ_INTEGRATION=1 to run against a dedicated RabbitMQ environment")
	}
	port, _ := strconv.Atoi(defaultEnv("CASINO_RABBITMQ_PORT", "5672"))
	managementPort, _ := strconv.Atoi(defaultEnv("CASINO_RABBITMQ_MANAGEMENT_PORT", "15672"))
	retryQueue := "integration.rabbit.retry"
	deadQueue := "integration.rabbit.failure"
	adapter := New(Config{
		Host: os.Getenv("CASINO_RABBITMQ_HOST"), Port: port, ManagementPort: managementPort,
		Username: os.Getenv("CASINO_RABBITMQ_USERNAME"), Password: os.Getenv("CASINO_RABBITMQ_PASSWORD"), VHost: defaultEnv("CASINO_RABBITMQ_VHOST", "/"),
		TopicExchange: "event.topic", DirectExchange: "event.direct", FanoutExchange: "event.fanout", RetryExchange: "event.delay", DeadExchange: "event.dead",
		Prefetch: 2, Concurrency: 1, PublisherConfirm: true, PublisherTimeout: 5 * time.Second, ReconnectDelay: time.Second, MaxRetries: 1,
		Queues: []string{retryQueue, deadQueue}, Bindings: map[string][]string{retryQueue: {"integration.retry"}, deadQueue: {"integration.dead"}},
	})
	t.Cleanup(func() { cleanupIntegrationQueues(adapter.connectionURL(), retryQueue, deadQueue) })
	dispatcher := &integrationDispatcher{calls: map[string]int{}, businessWrites: map[string]int{}, alwaysFail: map[string]bool{deadQueue: true}}
	registrations := []eventbus.ConsumerRegistration{
		{Name: "integration.retry.v1", Queue: retryQueue, EventTypes: []string{"integration.retry.v1"}, RoutingKeys: []string{"integration.retry"}, Handler: func(eventbus.Context) error { return nil }},
		{Name: "integration.dead.v1", Queue: deadQueue, EventTypes: []string{"integration.dead.v1"}, RoutingKeys: []string{"integration.dead"}, Handler: func(eventbus.Context) error { return nil }},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err := adapter.Start(ctx, registrations, dispatcher); err != nil {
		t.Fatal(err)
	}
	publishIntegrationEvent(t, adapter, "integration-retry-event", "integration.retry.v1", "integration.retry")
	publishIntegrationEvent(t, adapter, "integration-dead-event", "integration.dead.v1", "integration.dead")
	waitFor(t, 15*time.Second, func() bool {
		dispatcher.mu.Lock()
		defer dispatcher.mu.Unlock()
		return dispatcher.businessWrites["integration-retry-event"] == 1 && dispatcher.calls[deadQueue] >= 2
	})
	dead, err := adapter.DeadMessages(ctx, deadQueue, 10)
	if err != nil || len(dead) == 0 || dead[0].EventID != "integration-dead-event" {
		t.Fatalf("dead message not retained: %#v err=%v", dead, err)
	}
	dispatcher.mu.Lock()
	dispatcher.alwaysFail[deadQueue] = false
	dispatcher.mu.Unlock()
	if err := adapter.ReplayDead(ctx, deadQueue, "integration-dead-event"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, 10*time.Second, func() bool {
		dispatcher.mu.Lock()
		defer dispatcher.mu.Unlock()
		return dispatcher.businessWrites["integration-dead-event"] == 1
	})

	adapter.connectionMu.Lock()
	if adapter.connection != nil {
		_ = adapter.connection.Close()
	}
	adapter.connectionMu.Unlock()
	publishIntegrationEvent(t, adapter, "integration-reconnect-event", "integration.retry.v1", "integration.retry")
	publishIntegrationEvent(t, adapter, "integration-reconnect-event", "integration.retry.v1", "integration.retry")
	waitFor(t, 10*time.Second, func() bool {
		dispatcher.mu.Lock()
		defer dispatcher.mu.Unlock()
		return dispatcher.businessWrites["integration-reconnect-event"] == 1
	})
	if err := adapter.Close(); err != nil {
		t.Fatal(err)
	}
}

func publishIntegrationEvent(t *testing.T, adapter *Adapter, eventID, eventType, routingKey string) {
	t.Helper()
	envelope := eventbus.EventEnvelope{
		EventID: eventID, EventType: eventType, EventVersion: 1, OccurredAt: time.Now(), Producer: "integration-test", Payload: map[string]any{}, Metadata: map[string]any{},
	}
	body, _ := json.Marshal(envelope)
	if err := adapter.Publish(context.Background(), eventbus.PublishMessage{
		EventID: eventID, RoutingKey: routingKey, Body: body, OccurredAt: envelope.OccurredAt,
		ContentType: "application/json", MessageType: eventType,
	}); err != nil {
		t.Fatal(err)
	}
}

func waitFor(t *testing.T, timeout time.Duration, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}

func cleanupIntegrationQueues(connectionURL string, queues ...string) {
	connection, err := amqp091.Dial(connectionURL)
	if err != nil {
		return
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		return
	}
	defer channel.Close()
	for _, queue := range queues {
		_, _ = channel.QueueDelete(queue, false, false, false)
		_, _ = channel.QueueDelete(queue+".dead", false, false, false)
		_, _ = channel.QueueDelete(queue+".retry.5s", false, false, false)
	}
}

func defaultEnv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
