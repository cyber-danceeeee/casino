# Game Aggregator

This backend now has a provider-independent game aggregation layer under `internal/game`.

## Architecture

- `internal/game/domain`: normalized provider, launch, player, and wallet callback contracts.
- `internal/game/provider`: outbound provider and inbound callback adapter registries.
- `internal/game/provider/softaggregator`: SoftAggregator-specific API client, DTOs, mapper, signature validation, and callback adapter.
- `internal/game/repository`: Gorm access for providers, normalized games, provider-player mappings, provider transactions, and sync status.
- `internal/game/service`: catalog sync, launch/player creation, and wallet callback accounting.
- `internal/handler/gameProviderAggregator.go`: thin Gin controller for sync, launch, demo launch, and callbacks.

Core business code depends on these interfaces:

```go
type GameProvider interface {
    Code() string
    GetGames(ctx context.Context, req GetGamesRequest) ([]Game, error)
    CreatePlayer(ctx context.Context, req CreatePlayerRequest) (*ProviderPlayer, error)
    LaunchGame(ctx context.Context, req LaunchGameRequest) (*LaunchGameResult, error)
    LaunchDemo(ctx context.Context, req LaunchDemoRequest) (*LaunchGameResult, error)
    GetProviderBalance(ctx context.Context) (*ProviderBalance, error)
}
```

Inbound wallet protocols are isolated by callback adapter:

```go
type ProviderCallbackAdapter interface {
    Code() string
    ValidateRequest(ctx context.Context, request ProviderCallbackRequest) error
    ParseCommand(ctx context.Context, request ProviderCallbackRequest) (*WalletCommand, error)
    BuildSuccessResponse(ctx context.Context, command WalletCommand, result WalletResult) (*ProviderCallbackResponse, error)
    BuildErrorResponse(ctx context.Context, request ProviderCallbackRequest, err error) *ProviderCallbackResponse
}
```

## Adding Another Provider

A future provider should only need:

```go
type NewProvider struct{}

func (p *NewProvider) Code() string
func (p *NewProvider) GetGames(...)
func (p *NewProvider) CreatePlayer(...)
func (p *NewProvider) LaunchGame(...)
func (p *NewProvider) LaunchDemo(...)
func (p *NewProvider) GetProviderBalance(...)
```

Then implement `ProviderCallbackAdapter`, register both implementations in `service.NewRuntime`, and add environment configuration. The generic callback controller, wallet callback service, wallet repository, ledger logic, and game transaction accounting should not need provider-specific changes.

## Game Synchronization

Local game catalog route for the frontend:

```bash
curl "http://localhost:8080/api/v1/games?category=live-casino&provider=softaggregator&page=1&pageSize=30"
```

The frontend must use this local catalog DTO. It hides provider `rawMetadata` and exposes only normalized fields such as `providerCode`, `providerGameId`, `categoryCode`, `providerCategoryCode`, `thumbnailUrl`, `providerEnabled`, `platformEnabled`, and `effectiveEnabled`.

Provider sync-and-inspect route:

```bash
curl "http://localhost:8080/api/v1/game/providers/softaggregator/gamelist"
```

This route syncs the provider catalog into the local database first, then returns the local normalized catalog. Use `?sync=false` to read the local cache without calling the upstream provider. It does not expose provider credentials or `rawMetadata`.

Admin route:

```bash
curl -X POST "http://localhost:8080/api/v1/admin/game/providers/softaggregator/sync" \
  -H "Authorization: Bearer fake-admin-token"
```

Sync inserts new games and updates provider-controlled catalog fields: name, slug, category, game type, image, demo/free-spins capability, raw metadata, and `last_synced_at`. It preserves local operational fields such as `status`, sorting, popularity, badge, and launch overrides for existing rows.

## Launch Flow

Authenticated route:

```bash
curl -X POST "http://localhost:8080/api/v1/games/softaggregator/studio%2FGame/launch?userId=1" \
  -H "Content-Type: application/json" \
  -d '{"currency":"MXN","language":"es","device":"mobile","return_url":"https://example.test/games"}'
```

Demo route:

```bash
curl -X POST "http://localhost:8080/api/v1/games/softaggregator/studio%2FGame/launch/demo?userId=1" \
  -H "Content-Type: application/json" \
  -d '{"currency":"MXN","language":"es","device":"mobile","return_url":"https://example.test/games"}'
```

Mexico defaults are `MXN`, `es`, and `MX`. Provider credentials never appear in the response.

## Wallet Callbacks

Generic routes:

```text
GET /api/v1/game/providers/{providerCode}/callback
GET /api/v1/game/providers/{providerCode}/callbacks/{operation}
```

Each adapter validates the provider protocol and maps into a normalized `WalletCommand`. `WalletCallbackService` resolves the provider player mapping, locks `game_user_wallets`, applies exact decimal accounting to `cash_balance`, inserts immutable `game_wallet_transactions`, and inserts normalized `game_provider_transactions`.

Idempotency is enforced by:

```text
provider_code + provider_transaction_id + operation
```

Exact duplicate successful callbacks return the stored resulting balance without applying the balance change again. Conflicting duplicates are rejected as `TRANSACTION_CONFLICT`.

## Database

The schema adds:

- `game_provider_players`: provider player mapping and provider-side player secret.
- `game_provider_transactions`: normalized callback audit/idempotency table.
- `game_provider_syncs`: last sync status and counts.
- `game_games` columns: `slug`, `category`, `supports_demo`, `supports_free_spins`, `raw_metadata`, `last_synced_at`, plus a provider-game unique key.

For existing databases, use the unified installer and `database/schema.sql` during deployment.

## Verification

Useful local commands:

```bash
go test ./internal/game/...
make test
make build
```

The repository had pre-existing generated DAO test failures before this integration; track those separately from the focused `internal/game` tests.
