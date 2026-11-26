# Game Casino Backend Tables

This schema is derived from the current Vue H5 business surface. Every table uses the `game_` prefix and includes:

- `id` bigint unsigned auto increment primary key
- `created_at`
- `updated_at`
- `deleted_at`

The SQL file is [game_schema.sql](/Users/wimmer/Documents/axbros/game_casino/database/mysql/game_schema.sql).

## CRUD Endpoint Convention

After generating CRUD with sponge, use a consistent REST prefix:

```text
GET    /api/v1/{resource}
GET    /api/v1/{resource}/{id}
POST   /api/v1/{resource}
PUT    /api/v1/{resource}/{id}
DELETE /api/v1/{resource}/{id}
```

For example, `game_users` maps to `/api/v1/users`.

## Account And Auth

| Table | Resource | Purpose |
|---|---|---|
| `game_users` | `/api/v1/users` | Member account, phone login, VIP level, invite code |
| `game_user_sessions` | `/api/v1/user-sessions` | Login sessions and token lifecycle |
| `game_sms_verification_codes` | `/api/v1/sms-verification-codes` | Login/register/bind phone verification codes |
| `game_user_settings` | `/api/v1/user-settings` | Language, music switch, push switch |
| `game_user_devices` | `/api/v1/user-devices` | H5/app device records and risk checks |

Business APIs to implement above CRUD:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/send-code`
- `GET /api/v1/me`
- `PUT /api/v1/me/settings`

## Wallet And Payments

| Table | Resource | Purpose |
|---|---|---|
| `game_user_wallets` | `/api/v1/user-wallets` | Cash, bonus, withdrawable, frozen balances |
| `game_withdraw_accounts` | `/api/v1/withdraw-accounts` | Bank card and e-wallet accounts |
| `game_payment_methods` | `/api/v1/payment-methods` | EasyPaisa, JazzCash, QR scan, etc. |
| `game_payment_channels` | `/api/v1/payment-channels` | Third-party payment channels |
| `game_deposit_orders` | `/api/v1/deposit-orders` | Deposit orders |
| `game_withdraw_orders` | `/api/v1/withdraw-orders` | Withdraw orders |
| `game_wallet_transactions` | `/api/v1/wallet-transactions` | Wallet statement records |
| `game_cash_audit_records` | `/api/v1/cash-audit-records` | Turnover/audit requirements |
| `game_bonus_records` | `/api/v1/bonus-records` | Bonus center records |

Business APIs to implement above CRUD:

- `POST /api/v1/wallet/deposit`
- `POST /api/v1/wallet/withdraw`
- `POST /api/v1/wallet/withdraw-accounts`
- `GET /api/v1/wallet/records?tab=deposit|withdraw|statement`
- `GET /api/v1/cash-center`
- `GET /api/v1/bonus-center`

## Game Lobby

| Table | Resource | Purpose |
|---|---|---|
| `game_game_categories` | `/api/v1/game-categories` | Hot, Slots, Chance, Sports |
| `game_game_providers` | `/api/v1/game-providers` | JILI, PG Soft, SABA, etc. |
| `game_games` | `/api/v1/games` | Game cards and launch metadata |
| `game_game_category_relations` | `/api/v1/game-category-relations` | Game to category mapping |
| `game_user_favorite_games` | `/api/v1/user-favorite-games` | Heart/favorite state |
| `game_game_launch_logs` | `/api/v1/game-launch-logs` | Game loading/enter/fail logs |
| `game_bet_records` | `/api/v1/bet-records` | Bet record page |
| `game_bet_reports` | `/api/v1/bet-reports` | Bet report tab aggregates |

Business APIs to implement above CRUD:

- `GET /api/v1/home`
- `GET /api/v1/games?category=hot|slots|chance|sports`
- `POST /api/v1/games/{id}/launch`
- `POST /api/v1/games/{id}/favorite`
- `DELETE /api/v1/games/{id}/favorite`

## Home Content

| Table | Resource | Purpose |
|---|---|---|
| `game_banners` | `/api/v1/banners` | Carousel banners and top download bar |
| `game_home_popup_dialogs` | `/api/v1/home-popup-dialogs` | Register gift, download app, first deposit guide |
| `game_home_feature_cards` | `/api/v1/home-feature-cards` | Invite Spin, Invite Chest, Rebate feature cards |
| `game_rebate_jackpots` | `/api/v1/rebate-jackpots` | Rebate Jackpot number and period |

## Offers And Activities

| Table | Resource | Purpose |
|---|---|---|
| `game_activity_events` | `/api/v1/activity-events` | Offers/event list and detail base config |
| `game_activity_assets` | `/api/v1/activity-assets` | Activity page images by key |
| `game_activity_rules` | `/api/v1/activity-rules` | Event rules dialog/content |
| `game_activity_reward_rules` | `/api/v1/activity-reward-rules` | Reward tiers and amount ranges |
| `game_user_activity_records` | `/api/v1/user-activity-records` | User claims/participation records |
| `game_user_checkins` | `/api/v1/user-checkins` | Daily sign-in records |
| `game_invitation_wheel_rounds` | `/api/v1/invitation-wheel-rounds` | Wheel countdown/current round |
| `game_invitation_wheel_prizes` | `/api/v1/invitation-wheel-prizes` | Wheel prize pool |
| `game_invitation_wheel_spins` | `/api/v1/invitation-wheel-spins` | Spin records |
| `game_invitation_chest_configs` | `/api/v1/invitation-chest-configs` | Chest tiers |
| `game_user_invitation_chests` | `/api/v1/user-invitation-chests` | User chest progress/opening |
| `game_red_packet_rounds` | `/api/v1/red-packet-rounds` | Red packet rain rounds |
| `game_red_packet_claims` | `/api/v1/red-packet-claims` | Red packet reward records |
| `game_welfare_cards` | `/api/v1/welfare-cards` | Benefit card configs |
| `game_user_welfare_cards` | `/api/v1/user-welfare-cards` | User activated cards |

Business APIs to implement above CRUD:

- `GET /api/v1/offers?tab=events|vip`
- `GET /api/v1/activities/{slug}`
- `POST /api/v1/activities/{slug}/claim`
- `POST /api/v1/activities/daily-sign-in/check-in`
- `POST /api/v1/activities/invitation-wheel/spin`
- `POST /api/v1/activities/red-packet-rain/claim`

## Share, Ranking, Agent

| Table | Resource | Purpose |
|---|---|---|
| `game_invites` | `/api/v1/invites` | Referral registration and validity |
| `game_agent_commission_rules` | `/api/v1/agent-commission-rules` | Agent commission tier rules |
| `game_agent_commission_records` | `/api/v1/agent-commission-records` | Agent commission settlements |
| `game_ranking_periods` | `/api/v1/ranking-periods` | Ranking cycle/date |
| `game_ranking_records` | `/api/v1/ranking-records` | Ranking rows and self rank |

Business APIs to implement above CRUD:

- `GET /api/v1/share?tab=social|ranking|agent`
- `POST /api/v1/share/copy-link`
- `GET /api/v1/share/ranking`
- `GET /api/v1/share/agent/overview`

## VIP

| Table | Resource | Purpose |
|---|---|---|
| `game_vip_levels` | `/api/v1/vip-levels` | VIP levels and benefits |
| `game_vip_requirements` | `/api/v1/vip-requirements` | Level-up and retain requirements |
| `game_user_vip_logs` | `/api/v1/user-vip-logs` | VIP upgrade/retain logs |

Business APIs to implement above CRUD:

- `GET /api/v1/vip`
- `POST /api/v1/vip/rewards/{type}/claim`

## Support, CMS, App

| Table | Resource | Purpose |
|---|---|---|
| `game_support_categories` | `/api/v1/support-categories` | FAQ categories |
| `game_support_questions` | `/api/v1/support-questions` | FAQ question/answer |
| `game_support_tickets` | `/api/v1/support-tickets` | Support message threads |
| `game_support_messages` | `/api/v1/support-messages` | Support ticket messages |
| `game_notices` | `/api/v1/notices` | Notice and message tab |
| `game_cms_pages` | `/api/v1/cms-pages` | About us, user agreement, rich pages |
| `game_app_versions` | `/api/v1/app-versions` | App download version config |
| `game_app_download_records` | `/api/v1/app-download-records` | Download app reward record |
| `game_system_configs` | `/api/v1/system-configs` | Brand/global settings |

Business APIs to implement above CRUD:

- `GET /api/v1/support?tab=faq|notice|message`
- `GET /api/v1/cms/{pageCode}`
- `GET /api/v1/app/latest`
- `POST /api/v1/app/download-reward/claim`

## Notes For Service Layer

CRUD generation is enough for admin/backoffice scaffolding. Player-facing APIs should not directly expose raw CRUD for money-changing tables. These operations must be transactional:

- deposit success callback: update `game_deposit_orders`, `game_user_wallets`, `game_wallet_transactions`, and possible activity rewards
- withdraw submit/review: update `game_withdraw_orders`, freeze/unfreeze balances, write wallet transactions
- bet settlement: write `game_bet_records`, update reports, update audit progress, update VIP progress
- activity claim/spin/red-packet: verify eligibility, write user record, write wallet transaction, issue audit/bonus record
- agent commission settlement: aggregate invitees and valid bets, then write commission records and wallet transactions
