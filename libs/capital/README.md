# capital

Go/Goose microservice for payments, accounts, wallets, and financial transactions.

## Overview

`capital` is the financial core. It manages accounts (wallets), payments, vouchers, subscription plans, rates, and usage records. The wallet resource provides a user-facing interface to account balances and transaction history.

## Endpoints

### Core CRUD

| Method | Path | Description |
|--------|------|-------------|
| GET/POST/PATCH/DELETE | `/accounts` | Account CRUD |
| GET/POST/PATCH/DELETE | `/payments` | Payment CRUD |
| GET/POST/PATCH/DELETE | `/paymentlogs` | PaymentLog CRUD |
| GET/POST/PATCH/DELETE | `/plans` | Plan CRUD |
| GET/POST/PATCH/DELETE | `/plantypes` | PlanType CRUD |
| GET/POST/PATCH/DELETE | `/providers` | Provider CRUD |
| GET/POST/PATCH/DELETE | `/rates` | Rate CRUD |
| GET/POST/PATCH/DELETE | `/transactions` | Transaction CRUD |
| GET/POST/PATCH/DELETE | `/usages` | Usage CRUD |
| GET/POST/PATCH/DELETE | `/vouchers` | Voucher CRUD |
| GET/POST/PATCH/DELETE | `/vouchertypes` | VoucherType CRUD |

### Wallet

| Method | Path | Description |
|--------|------|-------------|
| POST | `/wallet/init` | Find-or-create the current user's account |
| GET | `/wallet/detail` | Fetch account details |
| GET | `/wallet/balance` | Return bookBalance + availableBalance |
| POST | `/wallet/upgrade` | Update daily/monthly limits |
| GET | `/wallet/transactions` | Paginated transaction list |

## Entities

Account, Payment, PaymentLog, Plan, PlanType, Provider, Rate, Transaction, Usage, Voucher, VoucherType

## Module

```go
import "github.com/thescaffold/gox-apps-capital/app"

app.AppModule{}
```

## Development

```bash
go test ./...
go build ./...
```
