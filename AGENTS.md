# Agent instructions

Standing rules for Cursor agents working on vent.

## Demo server sanity check

On every change (GUI or otherwise), before you finish:

1. Start the example admin demo if it is not already running: `just gen` / migrate if needed, then `go run ./examples/basic/cmd/server` (or `just dev`) on port **8080**.
2. Open the app in the browser: `http://localhost:8080/admin/`.
3. Log in as a sanity check with `admin@vent.com` / `test_user`.
4. Leave the server running on this agent's desktop so Troy can keep testing.
5. Do not open a second PR just to start the server.

If the server or login fails, fix or report the failure; do not claim done while the demo is broken.

## No tests in examples

Never create, keep, or regenerate test files under `examples/` (including `examples/basic/` and any `*_test.go` there).

Example apps are demos only. Put tests in library/package paths outside `examples/` (for example module-root package tests).

If you find example-app tests on the current branch, delete them.
