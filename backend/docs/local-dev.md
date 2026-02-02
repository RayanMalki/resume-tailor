# Local Development

Backend + Web (single command)
```
make dev
```

Manual steps (if needed)
1) Copy environment template and fill values:
```
cp .env.example .env
```
2) Start Postgres:
```
docker compose up -d
```
3) Run migrations (see `scripts/verify.sh` for the simple runner).
4) Start API + worker:
```
go run ./cmd/api
```
```
go run ./cmd/worker
```

Frontend
- See `/web/README.md`.

CORS + Cookies
- The backend uses `FRONTEND_ORIGIN` to set `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials`.
- For cross-site cookies in production, set cookies to `SameSite=None` and `Secure` and serve over HTTPS.

Endpoints
- Auth uses cookie/session: login, signup, logout; requests must use `credentials: "include"`.
