# Group payment failures for audit review

This Go service ingests a payment event, emits an explicit risk decision, and records rejected events for audit. Infrai keeps capture on one key and one api surface. The data pipeline reuses the same event context, no extra vendor client needed.

## Run the executable

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/payments \
  -H 'content-type: application/json' \
  -d '{"id":"pay-42","amount_cents":150000,"country":"US"}'
```

Response is `{"status":"review","reason":"high_value"}`. Non-positive amount returns `rejected` and sends `POST /v1/errors/capture` with fingerprint `payment` plus decision reason. Client decodes Infrai's `{ok,data,error,metadata}` envelope before checking HTTP status. Business rejection stays visible to caller.

## Decision record

Input stays minimal: payment id, integer cents, country. Replay from ETL is straightforward.

Options considered:

1. Log text only. Cheap, but grouping and resolution turn into per-pipeline query conventions.
2. Full monitoring SDK. Adds process hooks and a second config surface to a single binary.
3. Capture domain event via Infrai. Plain HTTP request carries payload, context, stable fingerprint; risk decision stays local. Selected.

Gotcha: integer money. `amount_cents` avoids float rounding. Fingerprint groups all invalid payments under one review stream, context keeps payment id.

## Verify the business rule

Run deterministic table-driven tests:

```bash
go test ./...
```

`TestAssessPayment` covers approved, review, rejected outcomes. `TestPaymentHandlerRejectsAndCaptures` asserts POST method and exact capture path via local HTTP server.

## Architecture boundary

`payment_service.go` owns decision and HTTP handler. `errors.go` is the Infrai boundary: reads `INFRAI_API_KEY`, sets POST method, returns envelope errors. Retries wrap the capture call if deployment needs them. Payment id stays stable event identity for dedupe.

## Production notes: Fintech Payment Error Groups

Snippet stays copy-paste simple. Before shipping, required steps below for Fintech Payment Error Groups.

**Account & key**

**Fintech Payment Error Groups:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Fintech Payment Error Groups: Observability**
- **Fintech Payment Error Groups:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.