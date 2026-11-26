# RabbitMQ Migration Audit

Audit date: 2026-07-20

This report was completed before the RabbitMQ implementation was changed. It records the queue-like behavior that existed in the repository and the migration decision for each path.

## Existing Event Infrastructure

| Area | Previous implementation | Risk | Migration decision |
| --- | --- | --- | --- |
| Event publication | Transactional MySQL `event_outbox` | Safe transaction boundary, but no external broker | Preserve as Transactional Outbox and relay to RabbitMQ |
| Dispatch | MySQL polling plus in-process handler calls | API instances execute consumers locally; no broker backpressure | Replace production dispatch with RabbitMQ adapter |
| Consumer delivery | `event_deliveries` | Stable consumer IDs and idempotent delivery already present | Preserve as consumer idempotency ledger |
| Retry | Fixed 30-59 second Outbox retry | No staged delay and no queue isolation | RabbitMQ TTL retry queues: 5s, 30s, 2m, 10m, 30m |
| Dead events | `event_outbox.status=DEAD` | No broker DLQ or retained broker payload | Keep Outbox DEAD for publish failures; add per-queue RabbitMQ DLQ for consume failures |
| Multi-worker claim | MySQL `FOR UPDATE SKIP LOCKED` | Correct for Outbox claims | Preserve for the Outbox relay |
| Ledger audit consumer | In-process | Low-risk read/write projection | Move to `audit.events` |
| Dashboard facts consumer | In-process | Low-risk projection | Move to `dashboard.events` |
| Promotion deposit consumer | In-process | Could execute twice after a crash | Move to `promotion.events`; retain business-table idempotency |
| Tenant onboarding email | In-process | Failed jobs remained in DB DEAD state | Move to `notification.events` |
| Push notifications | Polling worker executed FCM directly | Bypassed Event Bus | Poller now only schedules due jobs; RabbitMQ consumer sends notifications |
| Support exports | Detached goroutine | Lost on process restart | Move to `system.jobs` |
| Report exports | Detached goroutine | Lost on process restart | Move to `system.jobs` |
| Provider sync | Detached goroutine | Lost on process restart; no broker backpressure | Move to `provider.events` |
| Backup tasks | Detached goroutine | Lost on process restart | Move to `system.jobs` |
| Scheduled task execution | Scheduler called detached goroutine | Lost on process restart | Scheduler publishes; `system.jobs` executes |
| `game_system_retry_records` | Database operation record | Not a message queue | Retain as operation audit/history, not as MQ transport |
| Redis | Cache and scoped keys | Not used as a queue | No transport migration required |
| WebSocket/MQTT | No server-side queue implementation found | None | No migration required |
| H5 and Android projects | No server message queue implementation found | None | No migration required |

## Existing Consumers Found

- `audit.ledger_transaction.v1`
- `dashboard.daily_stats.v1`
- `promotion.deposit_trigger.v1`
- `notification.tenant_onboarding_email.v1`

Production migration also adds stable consumers for support export, report export, Provider sync, backup, scheduled task execution, and push notification jobs.

## Important Boundary

The Transactional Outbox remains intentional. Business transactions do not publish directly to RabbitMQ because a database commit and an AMQP publish cannot be made atomic without a distributed transaction. The final path is:

```text
business database transaction
  -> event_outbox in the same transaction
  -> short Outbox relay claim
  -> RabbitMQ publish with publisher confirm
  -> mark Outbox PUBLISHED
  -> manual-ack consumer
  -> event_deliveries plus consumer business idempotency
```

Business modules continue to depend only on `eventbus.EventPublisher`, `eventbus.EventConsumer`, `eventbus.EventDispatcher`, and `eventbus.QueueManager`. Only `internal/eventbus/rabbitmq` imports the AMQP client.
