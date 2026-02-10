# CLAUDE.md — Faktory

Faktory is a language-agnostic background job server (fork of [contribsys/faktory](https://github.com/contribsys/faktory)) maintained by Hunter.io. Workers in any language connect via TCP to push, fetch, acknowledge, and fail jobs. It uses an embedded Redis instance for persistence.

## Quick Reference

All commands run inside Docker — no local Go installation required.

| What | Command |
|---|---|
| Run tests | `make test` |
| Build production image | `make build` |
| Run Faktory | `make run` |
| Format code | `make fmt` (modifies local files via volume mount) |
| Generate webui templates | `make generate` (copies generated files locally) |
| Coverage report | `make cover` |
| Verify version sync | `make version_check` |

## Architecture Overview

```
Clients ──TCP:7419──▶ Server ──▶ Manager ──▶ Storage (Redis)
                        │
                        └──HTTP:7420──▶ WebUI
```

- **cmd/faktory/daemon.go** — entry point; boots server, registers WebUI subsystem, handles signals.
- **server/** — TCP listener, connection lifecycle, RESP protocol, worker heartbeat tracking, recurring task runner.
- **manager/** — core business logic: Push/Fetch/Ack/Fail, retry policy, reservation tracking, middleware chain.
- **storage/** — Redis-backed queues, sorted sets (retries, scheduled, dead), history/stats.
- **webui/** — HTTP dashboard (ego templates compiled to Go), CSRF-protected, serves on a separate port.
- **client/** — pure-Go TCP client library implementing the Faktory Work Protocol (FWP).
- **cli/** — CLI flag parsing, config loading (TOML), signal handling.
- **util/** — logging helpers, platform-specific utilities.

## Language & Dependencies

- **Go 1.25** (module: `github.com/hunter-io/faktory`)
- Key deps: `go-redis/redis`, `apex/log`, `BurntSushi/toml`, `justinas/nosurf`, `stretchr/testify`
- WebUI templates use **ego** (compiled at build time via `go generate`)
- Static assets embedded via **`//go:embed`**

## Project Layout

```
cmd/faktory/       Entry point (daemon.go)
server/            TCP server, connections, commands, workers, config
manager/           Job lifecycle (push/fetch/ack/fail), retries, middleware
storage/           Redis backend, queues, sorted sets, history
webui/             HTTP dashboard, ego templates (.ego), static assets
client/            Go client library, job struct, protocol
cli/               CLI setup, flag parsing, security
util/              Logging, platform helpers
test/              Integration & load tests
docs/              Protocol spec, release checklist
packaging/         RPM/DEB packaging scripts, systemd/upstart configs
```

## Code Conventions

- **Naming**: PascalCase for exported types/funcs, camelCase for unexported. Factory functions use `New*` prefix (e.g. `NewServer`, `NewManager`).
- **Error handling**: Always return `error`; no panics except on critical boot failures.
- **Concurrency**: One goroutine per client connection; `sync.Mutex`/`RWMutex` for shared state.
- **Interfaces**: Core abstractions (`Store`, `Queue`, `SortedSet`, `Manager`) are interfaces for testability.
- **Logging**: Use `util.Infof()`, `util.Warnf()`, `util.Debugf()`, `util.Error()` — not `fmt.Print` or `log.*`.
- **Tests**: `*_test.go` files live alongside their source in the same package. Tests run exclusively in Docker via `make test` (uses `Dockerfile.test`).

## Configuration

- TOML-based config files in `~/.faktory/` (dev) or `/etc/faktory/` (prod).
- Environment: `FAKTORY_SKIP_PASSWORD=true` disables auth (dev only).
- Default ports: TCP **7419**, WebUI **7420**.
- See `example/config.toml` for queue backpressure and TLS settings.

## CI/CD

- **GitHub Actions** (`.github/workflows/docker-publish.yml`): Tests on push to `master` / tags via `make test` (Docker), then builds & pushes Docker image to `ghcr.io/hunter-io/faktory`.

## Docker

All development commands run inside Docker for reproducibility. No local Go toolchain is needed.

- **`Dockerfile`** — production multi-stage build: Go 1.25 builder → Alpine 3.21 runtime with Redis.
- **`Dockerfile.test`** — Go 1.25 + Redis + code generation tools (`ego`). Used by `make test`, `make generate`, and `make cover`.
- Images are built and pushed to `ghcr.io/hunter-io/faktory` by GitHub Actions on push to `master` / tags.

## Version

Current version is defined in two places that must stay in sync:
- `Makefile` → `VERSION=0.15.0`
- `client/faktory.go` → `Version = "0.15.0"`

Run `make version_check` to verify they match.
