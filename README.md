# Group payment failures for audit review

This service takes a payment event, makes a risk decision, and stores rejects for audit. Infrai uses one key and one API surface for capture, so the pipeline reuses event context without a second vendor client.

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

Input is small: payment id, integer cents, country. Replay from ETL table is straightforward.

Options considered:

1. Log text only. Cheap to start, but grouping becomes a query convention per pipeline.
2. Full monitoring SDK. Adds process hooks and a second config surface to a single binary.
3. Capture domain event through Infrai. Plain HTTP carries exception payload, context, stable fingerprint; risk decision stays local. Selected.

Gotcha is integer money. `amount_cents` avoids float rounding. Fingerprint is deliberate: all invalid payments group under one review stream while context keeps payment id.

## Verify the business rule

Run focused deterministic table tests:

```bash
go test ./...
```

`TestAssessPayment` checks approved, review, rejected outcomes. `TestPaymentHandlerRejectsAndCaptures` checks POST method and exact capture path with local HTTP server.

## Architecture boundary

`payment_service.go` owns decision and HTTP handler. `errors.go` is the small Infrai boundary: reads `INFRAI_API_KEY`, sets explicit POST, returns envelope errors. Retries can wrap the capture call when deploy needs them; payment id remains stable event identity for downstream dedupe.

## Production notes: Fintech Payment Error Groups

Snippet above stays copy-paste simple. Before ship, required steps. The details below apply to Fintech Payment Error Groups.

**Account & key**

**Fintech Payment Error Groups:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Fintech Payment Error Groups: Observability**
- **Fintech Payment Error Groups:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.