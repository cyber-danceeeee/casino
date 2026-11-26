# SoftAggregator Provider

SoftAggregator configuration is environment-only. Never commit real credentials or callback salt values.

## Environment

```env
GAME_SOFTAGGREGATOR_ENABLED=true
GAME_SOFTAGGREGATOR_API_BASE_URL=https://api.softaggregator.com
GAME_SOFTAGGREGATOR_API_LOGIN=your-api-login
GAME_SOFTAGGREGATOR_API_PASSWORD=your-api-password
GAME_SOFTAGGREGATOR_SALT_KEY=your-callback-salt
GAME_SOFTAGGREGATOR_CALLBACK_PATH=/api/v1/game/providers/softaggregator/callback
GAME_SOFTAGGREGATOR_CALLBACK_IP_ALLOWLIST=
GAME_SOFTAGGREGATOR_MAX_TIMESTAMP_DRIFT_SECONDS=30
GAME_SOFTAGGREGATOR_REQUEST_TIMEOUT_SECONDS=10
GAME_SOFTAGGREGATOR_CONNECT_TIMEOUT_SECONDS=3
GAME_SOFTAGGREGATOR_MAX_TRANSACTION_AMOUNT_MINOR=100000000
```

`GAME_SOFTAGGREGATOR_API_BASE_URL` may be set to `https://api.softaggregator.com`; the client appends the documented `/api/v1` path when it is absent.

## Outbound API

The adapter sends JSON `POST` requests to the SoftAggregator API endpoint. Each request includes `api_login`, `api_password`, and `method`.

Implemented methods:

- `getGameList`
- `createPlayer`
- `getGame`
- `getGameDemo`

`GetProviderBalance` returns `ErrNotSupported` because the public SoftAggregator documentation for this integration does not define a provider-balance method.

## Callback Setup

In the SoftAggregator operator dashboard, set the callback URL to:

```text
https://your-api.example.com/api/v1/game/providers/softaggregator/callback
```

Callbacks are GET requests and must always receive HTTP 200 with provider-specific JSON. The adapter validates:

- HTTP method
- source IP when `GAME_SOFTAGGREGATOR_CALLBACK_IP_ALLOWLIST` is configured
- required fields
- timestamp drift
- signature key
- action
- amount format and configured maximum

Signature:

```text
key = md5(timestamp + GAME_SOFTAGGREGATOR_SALT_KEY)
```

The comparison is constant-time. Callback `key` is redacted before raw payload audit storage.

## Callback Actions

SoftAggregator actions map to normalized wallet operations:

- `balance` -> `BALANCE`
- `debit` -> `DEBIT`
- `credit` -> `CREDIT`
- `refund` -> `CREDIT` with the callback amount
- `rollback` -> `ROLLBACK` if supplied with `related_call_id`

The public SoftAggregator docs also mention `rb=1` on debit/credit callbacks. Those callbacks are processed like normal debit/credit rows and remain idempotent by `call_id`.

For `type=bonus_fs` debit callbacks, the adapter maps the amount to zero because documented free-spin debit callbacks must not deduct player funds.

## Error Mapping

SoftAggregator response body:

```json
{"error":0,"balance":10000}
```

Internal errors map to provider codes:

- `0`: success
- `1`: insufficient balance, with current balance
- `2`: invalid signature, invalid request, player not found, provider disabled, currency mismatch, provider unavailable, internal error

No stack traces, SQL errors, provider credentials, callback signatures, or launch URLs are returned.

## Local Testing

Use fake values only:

```bash
export GAME_SOFTAGGREGATOR_ENABLED=true
export GAME_SOFTAGGREGATOR_API_BASE_URL=http://127.0.0.1:9999
export GAME_SOFTAGGREGATOR_API_LOGIN=fake-login
export GAME_SOFTAGGREGATOR_API_PASSWORD=fake-password
export GAME_SOFTAGGREGATOR_SALT_KEY=fake-salt

go test ./internal/game/...
```

Callback example with fake signature:

```bash
curl "http://localhost:8080/api/v1/game/providers/softaggregator/callback?username=sa_1&currency=MXN&action=balance&call_id=test-call&timestamp=1720000000&key=fake-md5"
```

Use a real timestamp and `md5(timestamp + fake-salt)` when testing against a local database.
