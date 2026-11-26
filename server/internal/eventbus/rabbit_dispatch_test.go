package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"casino/internal/content"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type recordingPublisher struct {
	mu       sync.Mutex
	messages []PublishMessage
	err      error
}

func (p *recordingPublisher) Publish(_ context.Context, message PublishMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	p.messages = append(p.messages, message)
	return nil
}

func (p *recordingPublisher) Available(context.Context) error { return p.err }

func TestRelayPendingUsesStandardEnvelopeAndPublisherConfirmation(t *testing.T) {
	db := rabbitDispatchTestDB(t)
	service := New(db)
	record, err := service.Publish(context.Background(), Event{
		Scope: content.Scope{SiteID: 9}, SiteID: 9,
		EventType: EventWalletBalanceChanged, AggregateType: "ledger_transaction", AggregateID: "42",
		Source: "ledger", Producer: "ledger", IdempotencyKey: "relay-envelope", RequestID: "request-1",
		TraceID: "trace-1", CorrelationID: "correlation-1", Payload: map[string]any{"amount": "10.00"}, Metadata: map[string]any{"operatorID": 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	publisher := &recordingPublisher{}
	summary, err := service.RelayPending(context.Background(), publisher, []string{EventWalletBalanceChanged, EventWhiteLabelConfigPublished}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Published != 1 || len(publisher.messages) != 1 {
		t.Fatalf("unexpected relay result: %#v messages=%d", summary, len(publisher.messages))
	}
	message := publisher.messages[0]
	if message.RoutingKey != "wallet.balance.changed" || message.OutboxID != record.ID {
		t.Fatalf("unexpected published message: %#v", message)
	}
	var envelope EventEnvelope
	if err := json.Unmarshal(message.Body, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.EventID != record.EventID || envelope.EventVersion != 1 || envelope.RequestID != "request-1" || envelope.Producer != "ledger" {
		t.Fatalf("standard envelope was not preserved: %#v", envelope)
	}
	if fmt.Sprint(envelope.Metadata["operatorID"]) != "5" {
		t.Fatalf("metadata missing from envelope: %#v", envelope.Metadata)
	}
	second, err := service.RelayPending(context.Background(), publisher, []string{EventWalletBalanceChanged, EventWhiteLabelConfigPublished}, 10)
	if err != nil || second.Published != 0 || len(publisher.messages) != 1 {
		t.Fatalf("published outbox row was relayed twice: %#v err=%v", second, err)
	}
}

func TestRelayPendingDoesNotMarkPublishedWhenAdapterFails(t *testing.T) {
	db := rabbitDispatchTestDB(t)
	service := NewWithOptions(db, Options{MaxAttempts: 3})
	record, err := service.Publish(context.Background(), Event{
		Scope: content.Scope{Platform: true}, EventType: EventWhiteLabelConfigPublished,
		AggregateType: "site", AggregateID: "8", Source: "test", IdempotencyKey: "relay-failure", Payload: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	publisher := &recordingPublisher{err: errors.New("publisher confirm timeout")}
	summary, err := service.RelayPending(context.Background(), publisher, []string{EventWalletBalanceChanged, EventWhiteLabelConfigPublished}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Failed != 1 {
		t.Fatalf("publish failure was not retained: %#v", summary)
	}
	stored, _, err := service.Get(context.Background(), record.ID, content.Scope{Platform: true})
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != StatusFailed || stored.PublishedAt != nil {
		t.Fatalf("failed publish was incorrectly committed: %#v", stored)
	}
}

func TestRabbitDispatchSkipsDuplicateConsumerBusinessCall(t *testing.T) {
	db := rabbitDispatchTestDB(t)
	service := New(db)
	record, err := service.Publish(context.Background(), Event{
		Scope: content.Scope{SiteID: 2}, SiteID: 2,
		EventType: EventWalletBalanceChanged, AggregateType: "wallet", AggregateID: "3", Source: "wallet",
		IdempotencyKey: "duplicate-consume", Payload: map[string]any{"amount": "1.00"},
	})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(record.Envelope)
	calls := 0
	registration := ConsumerRegistration{
		Name: "wallet.projection.v1", Queue: QueuePayment, EventTypes: []string{EventWalletBalanceChanged},
		Handler: func(Context) error { calls++; return nil },
	}
	message := DeliveryMessage{Queue: QueuePayment, RoutingKey: RoutingKey(record.EventType), Body: body, Headers: map[string]any{HeaderOutboxID: record.ID}}
	if err := service.Dispatch(context.Background(), message, registration); err != nil {
		t.Fatal(err)
	}
	if err := service.Dispatch(context.Background(), message, registration); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("duplicate delivery executed business handler %d times", calls)
	}
}

func rabbitDispatchTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}
