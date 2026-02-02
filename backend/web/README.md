# Resume Tailor Web

Prerequisites
- Node.js 18+
- Backend running (`go run ./cmd/api`, `go run ./cmd/worker`)

Environment
Create `.env.local`:
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

Run
```
npm install
npm run dev
```

Notes: CORS + cookies
- Backend must set `Access-Control-Allow-Origin` to the frontend origin and allow credentials.
- For cross-site cookies in production, use HTTPS + `SameSite=None; Secure` on session cookies.
