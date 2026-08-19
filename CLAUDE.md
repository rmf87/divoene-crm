# CLAUDE.md — Divoene CRM

Backend (Go/Gin/SQLite) + Admin SPA (React/Vite/Tailwind).

Extracted from monorepo per RFC_059.

## Commands

| Command | Does |
|---------|------|
| `npm install` | Install admin deps |
| `go mod download` | Install Go deps |
| `just build-server` | Build server Docker image |
| `just build-admin` | Build admin SPA |
| `just test` | Run Go tests with coverage |
| `just vet` | Run go vet |

## Structure

```
server/              # Go backend (Gin, DDD, SQLite)
├── cmd/server/      # CLI entrypoint (cobra)
├── internal/core/   # Domain entities + services (pure)
├── internal/infra/  # Database, auth, scheduler, API clients
├── handlers/        # Gin HTTP handlers
├── middleware/       # CORS, auth middleware
└── router/          # Route registration

packages/admin/      # React admin SPA (Vite + Tailwind)
deploy/docker/       # Dockerfile.server, compose.test.yml
```

## Stack

- Go 1.25+, Gin, SQLite (WAL)
- JWT auth (appleboy/gin-jwt/v2)
- Clicksign (contracts), OpenPix/Woovi (PIX), WhatsApp Business
- React 19, Vite 8, Tailwind 4

## Testing

- `just test` — Go unit/integration tests
- `just vet` — Go static analysis
- `just test-venom` — Acceptance tests (requires compose stack)
