# P5-09: Usage wallet and overage billing

**Task ID:** P5-09  
**Owner:** Backend + Frontend  
**Phase:** 5 – Posts, listings, reports & integrations

## Objective

Implement a prepaid usage wallet for variable costs (WhatsApp conversations, SMS, AI credits). Agencies pre-load credit; overage is deducted from wallet balance. Agencies can set markup for resale to clients.

## Scope (agent-runnable)

- **Wallet model:**
  - `agency_wallets` table (id, agency_id, balance_cents, auto_reload_enabled, auto_reload_threshold_cents, auto_reload_amount_cents, updated_at).
  - `wallet_transactions` table (id, wallet_id, type `credit | debit`, amount_cents, description, reference_id, created_at).
- **Wallet operations:**
  - Credit: Payment via Stripe adds balance.
  - Debit: Usage (WhatsApp overage, AI credits) deducts balance.
  - Auto-reload: When balance drops below threshold, charge card and add credit.
- **Overage billing:**
  - When usage exceeds tier quota (e.g., WhatsApp convos > 2,000), check wallet balance.
  - If balance available, deduct and allow; if insufficient, block action or queue for later.
  - Log to `usage_events` with cost.
- **Agency markup:**
  - `agency_pricing` table or config (agency_id, resource_type, markup_percent).
  - When agency bills clients, apply markup to pass-through cost.
- **API:**
  - `GET /api/v1/wallet` – Get wallet balance and recent transactions.
  - `POST /api/v1/wallet/add-funds` – Add funds (create Stripe payment intent).
  - `PUT /api/v1/wallet/auto-reload` – Configure auto-reload.
- **UI:**
  - Wallet page: Balance, add funds, transaction history, auto-reload settings.
  - Usage alerts: Notify when balance is low.
- **Tier gating:** Usage wallet available at all tiers; markup settings at Growth+.

## References

- [Agency-Pricing-Tiers.md](../../Agency-Pricing-Tiers.md) – Usage Wallet section
- [SMB-Pricing-Tiers.md](../../SMB-Pricing-Tiers.md) – Overage billing
- [P3-07-usage-billing-stripe.md](P3-07-usage-billing-stripe.md) – Stripe integration

## Dependencies

- P3-07 (Stripe integration)
- P2-03 (WhatsApp usage)
- P4-05/P4-06 (AI usage)

## Files to create or modify

| File | Action |
|------|--------|
| `migrations/` | Create – `agency_wallets`, `wallet_transactions`, `agency_pricing` tables |
| `internal/domains/wallet/` | Create – Wallet operations, balance check, deduction |
| `internal/domains/billing/` | Extend – Stripe payment for wallet top-up, auto-reload |
| `api/openapi.yaml` | Add – Wallet endpoints |
| Usage check hooks | Extend – Before overage action, check and deduct wallet |
| Portal: billing/wallet | Create – Wallet balance, add funds, transaction history, auto-reload config |

## Acceptance criteria

- [ ] Agency wallet tracks balance; can add funds via Stripe.
- [ ] Overage usage deducts from wallet balance.
- [ ] If wallet balance insufficient, action is blocked with clear message.
- [ ] Auto-reload charges card when balance drops below threshold.
- [ ] Wallet transactions are logged and visible in history.
- [ ] Agency can configure markup for pass-through billing to clients.
- [ ] Low balance triggers notification/alert.
