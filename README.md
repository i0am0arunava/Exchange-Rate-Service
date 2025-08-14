# Exchange Rate Service (Go)

Production-grade solution for the assignment. Implements:
- REST API to convert amounts, fetch latest rates, and retrieve historical rates.
- In-memory cache with concurrency safety and singleflight to avoid duplicate fetches.
- Hourly refresh of latest rates.
- 90‑day history limit with validation.
- Graceful error handling and health check.
- Prometheus `/metrics`.

**Supported currencies:** USD, INR, EUR, JPY, GBP.

## Quickstart

### Run locally
```bash
# from project root
go run ./cmd/server
```

### With Docker
```bash
docker build -t exchange-rate-service:latest .
docker run -p 8080:8080 --rm exchange-rate-service:latest
```

### Configuration (env vars)
| Var | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `EXTERNAL_BASE_URL` | `https://api.exchangerate.host` | Third‑party API |
| `REFRESH_INTERVAL` | `1h` | How often to refresh "latest" rates |
| `HISTORY_LOOKBACK_DAYS` | `90` | Max historical lookback |
| `SUPPORTED_CURRENCIES` | `USD,INR,EUR,JPY,GBP` | Whitelist |
| `SERVE_STALE_WHILE_REFRESHING` | `true` | Serve stale latest rates while background refresh runs |

## API

### Convert
```
GET /convert?from=USD&to=INR&amount=100&date=2025-01-01
```
- `date` optional (uses “latest” if omitted).
- Response:
```json
{
  "from":"USD","to":"INR","amount":100,
  "rate":83.125,"converted":8312.5,
  "date":"2025-01-01","stale":false
}
```

### Latest rates
```
GET /rates/latest?base=USD&symbols=INR,EUR
```
Returns `{ "date":"2025-08-13","base":"USD","rates":{"INR":83.12,"EUR":0.92}, "stale":false }`

### Historical (max 90 days)
```
GET /rates/history?base=USD&symbol=INR&start=2025-05-01&end=2025-06-01
```
Returns `{ "base":"USD","symbol":"INR","series":[{"date":"2025-05-01","rate":82.9}, ...] }`

### Health
```
GET /healthz
```

### Metrics (Prometheus)
```
GET /metrics
```

## Design choices
- Store rates normalized by base **USD** for all supported currencies. Cross rates computed as `rateUSD[to] / rateUSD[from]`.
- Concurrency: RWMutex protects cache; `singleflight.Group` deduplicates in‑flight fetches per date.
- Scheduler refreshes latest rates every hour. If external API fails we keep serving last good snapshot and mark responses as `"stale": true`.
- Cache prunes entries older than the configured lookback.

## Tests
Run:
```bash
go test ./...
```
Covers core conversion logic, cache concurrency, 90‑day validation, and handler happy/edge paths with a mock third‑party client.

## Assumptions
- Future dates are invalid (400).
- Amount defaults to 1 if omitted.
- Only whitelisted currencies are served; codes are case-insensitive.
- For history requests we fetch the whole range in one call and fill the cache.
```

