# Design note

Identities and endpoints are separate because they answer different questions.
`node-a` says which participant this is; `127.0.0.1:8081` says where to reach it.
Two nodes share one codebase and are told apart by configuration, so identity
cannot be derived from an address that changes when the same node runs
elsewhere.

Validation is separate from binding: `Validate` inspects values and opens no
socket, so a valid address that is already in use still passes it and the
listen error surfaces later, from the component that owns the listener.
Detecting that in `Validate` would mean binding to find out, defeating the point
of checking cheaply first.

`/healthz` claims only that this process is answering, and which node it is. It
does not show a peer is reachable. Keeping the claim narrow stops a passing
check being mistaken for a working system.
