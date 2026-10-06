# Implementation log

## Task 1.1 — reproducible start

Ran `go version`, `git --version` and `go env GOMOD`; confirmed `GOMOD` points
at `source/go.mod`. `go build ./...` and the smoke tests succeed.

"Compiles" and "implements the behaviour" are different: an untouched scaffold
builds while still returning `501 TODO W01_HEALTH`, so a green build alone
proves nothing about the checkpoint. The first genuine failure observed was
`/healthz` answering 501 from the placeholder; the fix was implementing the
handler, and the test that moved from fail to pass was
`TestWeek01_Health`.

## Task 1.2 — validation

Rules implemented in `Config.Validate`: node id present and well formed, HTTP
address loopback with a port from 1 to 65535, shutdown timeout greater than
zero and no more than 30 seconds, and the optional peer address either empty or
a valid loopback address. The stage gate is why an empty optional field is
accepted now.

Errors name the field and the expectation, so a bad setting says what to
change. Port 0 is refused by `ValidID`/`ValidAddress` for the CLI: it is
meaningful as a test fixture, where the real port is read back afterwards, but
a node started with it would listen somewhere unpredictable.

## Task 1.3 — health

`Health(nodeID)` returns a handler that allows only GET, sets `Allow: GET` and
405 otherwise, and answers 200 with `node_id` and `status`. It uses the shared
`WriteJSON` and `ErrorJSON` helpers so headers are set before the status and
the status before the body, which is the order net/http requires.

## Task 1.4 — two nodes

`week01/scripts/run-node-a.sh` and `run-node-b.sh` build the binary and then run it,
rather than using `go run`. With `go run` the process receiving Ctrl-C is the
go tool, not the node, so the shutdown being demonstrated would not be the
node's.

## Task 1.5 — failure paths

Recorded above in the evidence record. The listen failure is reported by
`app.New`, which binds explicitly instead of using `ListenAndServe`, so a busy
port is an error before the node claims to have started.

Shutdown trace: interrupt -> context cancellation -> `Run`'s select takes the
`ctx.Done()` branch -> `shutdown` drains with a fresh, bounded context ->
listener closed. The context is new because the one passed to `Run` is already
cancelled; reusing it would abort the drain immediately and make the graceful
path do nothing.

Ownership: `app` owns the listener, because it called `net.Listen`. It also
owns shutdown. A handler must not call `os.Exit` on a bad request: it cannot
know whether the process should end, and one malformed request from one client
is no reason to terminate every other in-flight request.
