# SMB Client Creation, Lead vs Active, and Per-Client Layers

**Plan type:** Additional (SMB creation flow, lead vs active lifecycle, per-client layers)  
**Owner:** Full-stack  
**References:** [FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md](../FARO_HQ_Strategic_Roadmap_CTO_CPO_Review.md), [P1-03-client-location-apis-ui.md](P1-03-client-location-apis-ui.md)

---

## Current state

- **Clients** are created with `CreateClientHandler` via `POST /api/v1/tenants/{id}/clients` with only `name`, `slug`, `tier`. No Google search, no address/website/social, no "lead" state.
- **Client model** ([internal/domains/tenants/domain/model/client.go](../../internal/domains/tenants/domain/model/client.go)): id, agencyID, name, slug, tier (starter/growth/scale), status (active/inactive/suspended). No lifecycle "lead", no website/social.
- **Location model** ([internal/domains/tenants/domain/model/location.go](../../internal/domains/tenants/domain/model/location.go)): clientID, name, address (JSONB), phone, businessHours, categories. No `gbp_place_id`, no website.
- **Strategic Roadmap:** $49/client add-on; agency tiers cap client count. No per-client "layer" feature matrix today.

---

## 1. Client creation flow: Google search + manual fallback

**Goal:** Easy SMB creation: search by name or address via Google, prefill data; if not found, manual entry with address (Google-assisted) plus name, website, social links.

**Assumption:** "Google search for businesses" and "location search by address" use **Google Places API** (Text Search / Find Place / Place Details). If you use another API (e.g. SerpAPI), keep the same endpoint contract and swap the integration.

**Backend (farohq-core-app):**

- New domain or module (e.g. `internal/domains/smbsearch/` or under tenants): Google Places client (API key from config).
  - Search by query: Text Search or Find Place (business name + city) → list of places (place_id, name, formatted_address, phone, etc.).
  - Search by address: Geocoding or Find Place with address → same shape.
- New endpoint: `GET /api/v1/smb/search?q=...` or `POST /api/v1/smb/search` with body `{ "query": "..." }` or `{ "address": "..." }`. Returns list of candidate places. Auth + tenant context.
- Optional: `GET /api/v1/smb/place-details?place_id=...` for full details before create.
- Create client flow (extended):
  - **From place_id:** Body includes `place_id`. Backend calls Place Details, creates **client** (name, slug) + **one location** with name, address, phone, **gbp_place_id**. Client created as **lead** by default.
  - **Manual:** No place_id; body has `name`, `address`, `phone`, `website`, `social_links`. Create client + one location; gbp_place_id null until GBP connected later.

**Portal (farohq-portal):**

- "Add client" / "Add SMB" flow:
  1. Search step: input (business name or address) → call search API → show results; user picks one or "My business isn't listed".
  2. If listed: prefilled form (name, address, phone); submit with `place_id`.
  3. If not listed: form with name, address (autocomplete), phone, website, social links; submit without place_id.
- Single create call (client + location); success → client detail or list.

**Acceptance criteria (creation):**

- [ ] Agency can search by business name or address and get a list of Google places.
- [ ] Selecting a place prefills client + location and stores gbp_place_id on location.
- [ ] "Not listed" path allows manual entry with address (Google-assisted), name, website, social links; client and one location created.
- [ ] Created SMB is easy (minimal required fields); by default created as **lead**.

---

## 2. Lead vs active: no features until activated with a layer

**Goal:** SMBs can exist as **leads** (no feature access). When the agency "activates" the business, they choose a **layer**; only then do features apply.

**Lifecycle:**

- **lead** – No inbox, reviews, AI, or reports. Shown as "Lead" with "Activate" action.
- **active** – Activated with a layer. Feature set according to layer (section 4).

**Backend:**

- Add **lifecycle** on client: `lead` | `active` (default `lead`). Or use `status = 'lead'` and "activate" = set status to active + set layer.
- **Activate:** e.g. `PATCH /api/v1/tenants/{id}/clients/{clientId}/activate` with body `{ "layer": "basic" | "growth" | "premium" }`. Validates limits; sets lifecycle = active and client.layer; may trigger billing.
- **Feature gating:** Inbox, reviews, diagnostics, AI, reports check: client must be **active** and **layer** must include that feature. Lead or missing feature → 403 or "Upgrade this client".

**Portal:**

- Client list/card: show "Lead" vs "Active" and layer name.
- Lead clients: "Activate" button → modal to select layer → confirm → call activate API.
- Active clients: show feature tabs; hide/disable features not in layer or show upgrade CTA.

**Acceptance criteria (lead vs active):**

- [ ] New clients created as lead; no access to inbox, reviews, AI, reports.
- [ ] "Activate" requires selecting a layer; after activation, features gated by layer.
- [ ] Requests for features for a lead client (or feature not in layer) return 403 or equivalent.

---

## 3. Schema and data model changes

**Clients table (migration):**

- Add `lifecycle` TEXT: `lead` | `active` (default `lead`).
- Add `layer` TEXT (nullable): e.g. `basic` | `growth` | `premium` (set when lifecycle = active).
- Add optional `website` TEXT, `social_links` JSONB.
- Keep `tier` for agency level; **layer** is source of truth for per-client features.

**Locations table (migration):**

- Add `gbp_place_id` TEXT (nullable) – Google Place ID for GBP OAuth/sync later.

**Layer feature matrix (code or table):**

- Define which features each layer has (e.g. `ai_responses`, `ai_automations`, `messages_included`, `reports_included`, `tracking_reports`, `comparison_reports`). Use for feature gating.

**Files:** [migrations/000004_agency_hierarchy.up.sql](../../migrations/000004_agency_hierarchy.up.sql), [internal/domains/tenants/domain/model/client.go](../../internal/domains/tenants/domain/model/client.go), [location.go](../../internal/domains/tenants/domain/model/location.go).

---

## 4. Per-client layers and feature matrix (Local Dominator / Local Falcon style)

**Goal:** When activating, agency selects a **layer**. Each layer unlocks a set of features; higher layers add AI, automations, messages, reports, and tracking/comparison for ROI proof.

**Suggested layers:**

| Layer     | Features (example)                                                                 |
|----------|-------------------------------------------------------------------------------------|
| **Basic** | Presence/reviews read-only; basic diagnostic; no AI, no inbox, no comparison      |
| **Growth** | + Inbox (messages); + AI reply suggestions; + Reports included                    |
| **Premium** | + AI automations; + Tracking and comparison reports; ROI proof                    |

**Tracking and comparison (Local Dominator / Local Falcon style):**

- **Tracking:** Snapshots over time (review count, rating, presence score, Maps position/relevance) per location or client.
- **Comparison:** Before vs after, or this month vs last month for reviews, presence, Maps visibility.
- **ROI proof:** "We increased reviews from X to Y", "presence score from A to B", "better positioning in Google Maps search".

**Backend:** Feature gate helper (client + feature → allowed?); tracking domain (snapshots); optional billing mapping (layer → price).

**Portal:** Activate modal with layer options and feature list; client detail shows layer and features; tracking/comparison sections only if layer includes them.

**Acceptance criteria (layers):**

- [ ] Activate flow shows layer options with feature list (and optionally price).
- [ ] Feature access (inbox, AI, reports, tracking/comparison) enforced by client layer in API and UI.
- [ ] Tracking and comparison reports (or first version) show improvement over time and support ROI narrative.

---

## 5. Implementation order (suggested)

1. **Schema:** Migration for clients (lifecycle, layer, website, social_links) and locations (gbp_place_id). Update client and location models and repositories.
2. **Google Places integration:** SMB search module, env for API key, search + optional place-details endpoints. Create client + first location from place_id or manual payload.
3. **Lead vs active:** New clients as lead by default; activate endpoint; feature gate "client active + layer includes feature" on feature routes.
4. **Layer definitions and feature gating:** Define layers and feature matrix; gate inbox, reviews, AI, reports by client layer.
5. **Portal:** Add-client flow (search + manual); activate modal (layer selection); client list/detail (lead vs active, layer); hide/disable features by layer.
6. **Tracking and comparison:** Snapshot job or events; storage for metrics over time; comparison/ROI views (can be Phase 2 for MVP).

---

## 6. Open points

- **Google API:** Confirm Places API (Text Search / Find Place / Place Details). If different (e.g. SerpAPI), keep same endpoint contract.
- **Layer names and prices:** Align with $49/client; add Basic (cheaper) and Premium (higher) or match brand (e.g. Local Dominator tier names).
- **First location:** Recommendation: always create one location per client on create; "Add another location" later.
- **Billing:** Should activate create a Stripe subscription item or usage record for that client + layer? If yes, activate endpoint (or billing service) updates subscription after layer is set.
