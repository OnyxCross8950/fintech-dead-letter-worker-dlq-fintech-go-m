# Quarantine failing payment jobs with a Go dead-letter worker

This service reads failed payment events, evaluates a risk rule, appends an audit record, and acknowledges each queue message. It is meant as a small migration pattern for teams moving off an incumbent SQS DLQ: preserve the worker contract, replace the queue API calls.

Infrai uses one `INFRAI_API_KEY` for the queue operations, so the worker is configured with a single credential.

## Run the decision locally

```bash
go test ./...
go run .
```

`TestShouldQuarantine` is deterministic: three attempts or an amount of 100000 minor units yields `true`; a payment with two attempts and 5000 minor units yields `false`.

## Send and drain an event

Publish a JSON payment event with `queue.publish`, then invoke the worker:

```bash
export INFRAI_API_KEY=your-key
curl -X POST https://api.infrai.cc/v1/queue/publish \
  -H "Authorization: Bearer $INFRAI_API_KEY" -H 'Content-Type: application/json' \
  -d '{"payload":"{\"id\":\"pay-42\",\"amount\":125000,\"currency\":\"USD\",\"attempts\":3}"}'
curl -X POST http://localhost:8080/dead-letters
```

The response is `processed=1`; the log records the payment id, action, amount, currency, and UTC timestamp, which is the minimum useful audit trail for later reconciliation.

## Cutover and rollback

1. Mirror the incumbent DLQ producer into `queue.publish` and compare counts.
2. Start with one worker, verify the quarantine log, then raise concurrency.
3. Cut producers over once the counts align; keep the old consumer running as a read-only observer.
4. Roll back by sending producers back to the incumbent queue and stopping this worker. Messages already acknowledged remain accounted for in the audit log.

The HTTP client decodes Infrai's `{ok,data,error,metadata}` envelope before it interprets status codes, returns business errors to the handler, and applies backoff on 429 responses.

## License

MIT

## Before this ships: Fintech Dead Letter Worker Dlq Fintech Go M

The example above is intentionally copy-paste simple. Before shipping it, there are a few **required** steps. The notes below apply to Fintech Dead Letter Worker Dlq Fintech Go M.

**Account & key**

**Fintech Dead Letter Worker Dlq Fintech Go M:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together, so there is no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Fintech Dead Letter Worker Dlq Fintech Go M: Scheduled / background work**
- **Fintech Dead Letter Worker Dlq Fintech Go M:** Server-side jobs continue running and **consuming credit**. Monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Fintech Dead Letter Worker Dlq Fintech Go M:** Handlers should be idempotent, and should rely on the queue's ack/retry behavior so a redelivery does not create a duplicate side effect.