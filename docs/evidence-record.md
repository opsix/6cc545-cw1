# Week 1 evidence record

Environment: Go 1.24.5 (`.tools/go/bin/go version`), module
`example.edu/6cc545/cw1`.

## Commands and results

| Command | Result |
|---|---|
| `go build ./...` | succeeds |
| `go vet ./...` | clean |
| `gofmt -l .` | no output (all formatted) |
| `go test ./... -run '^TestSmoke_'` | pass |
| `go test ./... -count=1 -timeout 20s` | pass |
| `go run ./examples/firstfunction/main` | `lab-demo` (the brief's target) |

## Task evidence

**Task 1.2 — configuration.** `TestWeek01_Config` covers 19 cases
table-driven: valid input, empty id, id with a space, leading separator, digits
and separators accepted, missing port, wildcard address, port 0, port above
range, non-numeric port, `localhost`, zero timeout, negative timeout, exactly
`MaxTimeout`, beyond `MaxTimeout`, and the optional peer field empty, valid and
invalid.
`TestWeek01_ConfigIndependentOfBinding` records that a valid address may still
be in use.

**Task 1.3 — health endpoint.** `TestWeek01_Health` asserts status 200, the
JSON media type, and the decoded `node_id` and `status` values, and that the
body has exactly those two fields. `TestWeek01_HealthWrongMethod` checks POST,
PUT and DELETE each return 405 with `Allow: GET` and the shared error shape.

**Task 1.4 — two instances.** Node A on `127.0.0.1:8081` returned
`{"node_id":"node-a","status":"ok"}`; node B on `127.0.0.1:8082` returned
`{"node_id":"node-b","status":"ok"}`. The identities differ.

**Task 1.5 — startup and shutdown failures.**

| Attempt | Exit | Message |
|---|---|---|
| Second start on A's address | 1 | `listen tcp 127.0.0.1:8081: bind: address already in use` |
| Invalid id `bad id` | 1 | `node id "bad id" is not valid: ...` |
| Port `99999` | 1 | `http address "127.0.0.1:99999" is not valid: ...` |
| Unknown flag `-nonsense` | 2 | `flag provided but not defined: -nonsense` |

In every failure the original node kept answering. After Ctrl-C the client got
`connect: connection refused`, and restarting on the same address worked, which
shows the listener was released rather than leaked.

**Task 1.6 — own test.** `TestHealthIdentityComesFromArgument` builds the
handler for three identities and asserts each responds with its own, which
shows the identity belongs to the caller rather than to the handler.
