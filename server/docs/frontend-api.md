# H5 Frontend Business API

Base path: `/api/v1`

Generated CRUD route groups have been renamed from table-style paths to semantic paths. H5 business actions are implemented on the handler that owns the primary table for that feature.

## Auth and User

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/auth/login` | Login with phone/password. Creates a demo user if the phone does not exist, which keeps early H5 integration simple. |
| `POST` | `/auth/register` | Register with phone/password/inviteCode. |
| `POST` | `/auth/logout` | Logout placeholder. |
| `GET` | `/me` | Current user profile and wallet summary. |
| `GET` | `/profile` | Profile page aggregate: user, wallet, settings, withdraw accounts, quick menus. |
| `GET` | `/settings` | User settings. |
| `PUT` | `/settings` | Update language, music switch, push switch. |

User lookup during development accepts `userId` query, `X-User-ID`, `phone` query, or `X-User-Phone`.

## Home and Games

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/home` | Home aggregate: banners, categories, providers, feature cards, popups, rebate jackpot, social links. It does not include game catalog data. |
| `GET` | `/games?category=hot&type=slot&keyword=ace` | Local normalized game catalog for the lobby. |
| `POST` | `/games/:id/launch` | Create a launch log and return a launch URL/token. |
| `POST` | `/games/:id/favorite` | Toggle favorite status. |

## Offers and Activities

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/offers` | Events and VIP levels for the offers page. |
| `GET` | `/vip` | VIP levels and requirements. |
| `GET` | `/activities/:slug` | Activity detail: event metadata, assets, rules, reward rules, current state. |
| `POST` | `/activities/:slug/action` | Activity action placeholder such as claim/start/spin. |

## Wallet and Security

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/wallet/summary` | Wallet summary. |
| `GET` | `/wallet/payment-methods?type=deposit` | Deposit/withdraw payment methods and channels. |
| `GET` | `/wallet/deposit/config` | Active-country deposit channels, currency-specific amount presets, eligible vouchers and their exact reward tiers. Empty payment setup returns an empty methods/channels catalog without hiding presets or vouchers. `available=false` and `unavailableReason` describe incomplete voucher setup. |
| `POST` | `/wallet/deposit` | Create deposit order. Body: `amount`, `methodCode`, `channelCode`, optional `voucherCode`. Reward tiers and currency are revalidated on the server. |
| `POST` | `/wallet/withdraw` | Create withdraw order. Body: `amount`, `withdrawAccountId`. |
| `GET` | `/wallet/records?tab=deposit|withdraw|statement` | Wallet records by tab. |
| `GET` | `/security/bank-account` | Withdrawal accounts. |
| `POST` | `/security/bank-account` | Add withdrawal account. |

## Records, Support, CMS, Share

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/bet-records?tab=records|report&status=all&gameType=slot` | Bet records or aggregate reports. |
| `GET` | `/support` | FAQ categories, questions, and notices. |
| `GET` | `/cms/:pageCode` | Static pages such as `about-us`, `user-agreement`, `faq`. |
| `GET` | `/share?tab=social|ranking|agent` | Share page aggregate: invite link, ranking, commission records. |
| `GET` | `/app/version?platform=android` | Latest app version and download URL. |
