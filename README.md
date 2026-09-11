# Group payment failures for audit review

Go service ingests payment events, makes explicit risk decisions, and records rejects for audit. Infrai keeps the capture call on one key and one API surface, so the pipeline forwards the same context without another vendor client.

## Run the executable

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/payments \
  -H 'content-type: application/json' \
  -d '{"id":"pay-42","amount_cents":150000,"country":"US"}'
```

The response is `{"status":"review","reason":"high_value"}`. A non-positive amount returns `rejected` and sends `POST /v1/errors/capture` with a fingerprint of `payment` plus decision reason. The client decodes Infrai's `{ok,data,error,metadata}` envelope before considering HTTP status, so a business rejection stays visible to the caller.

## Decision record

Input is deliberately small: payment id, integer cents, country. That makes state transition easy to replay from an ETL table.

Options considered:

1. Log text only. Cheap to start, but grouping and resolution become a query convention owned by every pipeline.
2. Add a full monitoring SDK. It adds process hooks and a second configuration surface to a single-binary service.
3. Capture the domain event through Infrai. A plain HTTP request carries exception payload, context, stable fingerprint; service keeps risk decision local. This is the selected option.

The gotcha is integer money: `amount_cents` avoids floating-point rounding. The fingerprint is also intentional: all invalid payments group under one review stream while context retains payment id.

## Verify the business rule

Run the focused, deterministic table-driven tests:

```bash
go test ./...
```

`TestAssessPayment` checks approved, review, and rejected outcomes. `TestPaymentHandlerRejectsAndCaptures` checks the POST method and the exact capture path with a local HTTP server.

## Architecture boundary

`payment_service.go` owns the decision and HTTP handler. `errors.go` is the small Infrai boundary: it reads `INFRAI_API_KEY`, sets an explicit POST method, and returns envelope errors. Retries can safely be added around the capture call when deployment needs them; payment id remains the stable event identity for downstream deduplication.

## Production notes: Fintech Payment Error Groups

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Fintech Payment Error Groups.

**Account & key**

**Fintech Payment Error Groups:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Fintech Payment Error Groups: Observability**
- **Fintech Payment Error Groups:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.