# Week 1 node foundation

One node of a distributed application: a process with its own identity that
listens on a loopback address, reports its liveness, and stops cleanly when
asked. This is the first increment of the application that grows through
Week 5.

## Layout

| Path | Responsibility |
|---|---|
| `cmd/node/` | Parse flags, validate, hand over to the app |
| `cmd/request/` | HTTP probe, so the demo does not depend on curl |
| `internal/config/` | Settings and the rules for checking them |
| `internal/app/` | Owns the listener, the HTTP server and shutdown |
| `internal/httpapi/` | JSON helpers and the liveness handler |
| `examples/firstfunction/` | The worked example: returning a value and an error |

## Running

```bash
go build ./...

# Two independent nodes
sh scripts/run-node-a.sh     # 127.0.0.1:8081, identity node-a
sh scripts/run-node-b.sh     # 127.0.0.1:8082, identity node-b

# From another terminal
go run ./cmd/request -url http://127.0.0.1:8081/healthz
```

Ctrl-C stops a node: interrupt, then context cancellation, then a bounded
graceful shutdown, then the listener is released.

## Testing

```bash
go test ./... -count=1 -timeout 20s
go vet ./...
gofmt -l .
```

## What `/healthz` does not prove

It says this process is responding and which node it is. It does not say any
other node is reachable. See `docs/design-note.md`.

## Submitting this

Submission instructions, the evidence files, and a showcase script are in the
`submission/` folder beside this repository. It is deliberately outside the
repo so the guide is readable before the first push.
