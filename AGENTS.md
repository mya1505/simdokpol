# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` holds entry points: `cmd/main.go` for the app plus dev tools (e.g., `cmd/signer/`, `cmd/license-manager/`).
- `internal/` contains core Go packages: `controllers/`, `services/`, `repositories/`, `models/`, and `middleware/`.
- `migrations/` stores SQL migrations grouped by database (`sqlite/`, `mysql/`, `postgres/`).
- `web/` includes UI assets and server-side templates (`web/static/`, `web/templates/`, `web/assets.go`).
- `tests/e2e/` contains end-to-end tests that exercise the running server.

## Build, Test, and Development Commands
- `go mod tidy` refreshes Go module dependencies.
- `air` runs the dev server with live-reload (configured in `.air.toml`).
- `go build -ldflags "-s -w" -o simdokpol cmd/main.go` builds a production binary.
- `make package` opens the interactive build menu; `make windows-installer` / `make linux-installer` build platform installers.

## Coding Style & Naming Conventions
- Go code is formatted with `gofmt` (tabs for indentation); keep `go vet`-clean where possible.
- Packages are lower-case; exported identifiers use `CamelCase`.
- Tests use `_test.go` suffix; e2e tests live under `tests/e2e/`.
- Template and asset changes should stay in `web/templates/` and `web/static/`.

## Testing Guidelines
- Unit tests (if applicable): `go test ./...`.
- End-to-end tests: start the server (`air` or `go run ./cmd/main.go`), then run `go test ./tests/e2e -v`.
- E2E expects a fresh setup at `http://localhost:8080` and uses SQLite defaults.

## Commit & Pull Request Guidelines
- Commit messages are short and imperative; common prefixes include `fix:` and `update` (see recent history).
- PRs should include a clear summary, testing notes, and screenshots for UI changes.
- Link related issues or discussions when applicable.

## Security & Configuration Notes
- Avoid committing secrets; build-time secrets such as `APP_SECRET_KEY` are injected via environment variables.
- Local configuration is generated during setup; keep `.env` out of the repository.
