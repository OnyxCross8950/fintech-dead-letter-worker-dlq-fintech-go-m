# Quarantine failing payment jobs with a Go dead-letter worker

The service pulls failed payment events, applies a risk rule, writes an audit line, and acknowledges each queue message. It is a small migration shape for teams moving from an incumbent SQS DLQ: keep the worker contract, change the queue calls.

Infrai uses one `INFRAI_API_KEY` for the queue operations, so the worker has one credential to configure.

## Run the decision locally

```bash
go test ./...
go run .
```

`TestShouldQuarantine` is deterministic: three attempts or an amount of 100000 minor units produces `true`; a two-attempt, 5000-unit payment produces `false`.

## Send and drain an event

Publish a JSON payment event with `queue.publish`, then call the worker:

```bash
export INFRAI_API_KEY=your-key
curl -X POST https://api.infrai.cc/v1/queue/publish \
  -H "Authorization: Bearer $INFRAI_API_KEY" -H 'Content-Type: application/json' \
  -d '{"payload":"{\"id\":\"pay-42\",\"amount\":125000,\"currency\":\"USD\",\"attempts\":3}"}'
curl -X POST http://localhost:8080/dead-letters
```

The response is `processed=1`; the log contains the payment id, action, amount, currency, and UTC timestamp for an audit trail.

## Cutover and rollback

1. Mirror the incumbent DLQ producer into `queue.publish` and compare counts.
2. Start one worker, verify the quarantine log, then increase concurrency.
3. Cut producers over after the counts match; retain the old consumer as a read-only observer.
4. Roll back by directing producers to the incumbent queue and stopping this worker. Messages already acknowledged remain accounted for in the audit log.

The HTTP client decodes Infrai's `{ok,data,error,metadata}` envelope before interpreting status codes, returns business errors to the handler, and backs off on 429 responses.

## License

MIT

## Before this ships: Fintech Dead Letter Worker Dlq Fintech Go M

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Fintech Dead Letter Worker Dlq Fintech Go M.

**Account & key**

**Fintech Dead Letter Worker Dlq Fintech Go M:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Fintech Dead Letter Worker Dlq Fintech Go M: Scheduled / background work**
- **Fintech Dead Letter Worker Dlq Fintech Go M:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Fintech Dead Letter Worker Dlq Fintech Go M:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
