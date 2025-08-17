


## Basic setup

1. Start  API service:
   ```bash
   git clone https://github.com/i0am0arunava/Exchange-Rate-Service.git
   ```
   then 
   ```bash
   cd exchange-rate-service
   ```
   then
   ```bash
   docker compose up --build
   ```
   or
   ```bash
   docker compose up
   ```
## Unit Testing

Follow the steps below to run the unit tests for this project.

1. **Ensure Docker Container is Running**  
   Make sure your API service and any dependencies (e.g., Memcache) are running inside Docker in the background:  
   ```bash
   docker compose up 
   ```

2. **Install Dependencies**  
   Run the following command to install all Go module dependencies:  
   ```bash
   go mod tidy
   ```

3. **Run Tests**  
   Execute the unit tests for the handler package:  
   ```bash
   go test -v ./internal/handler
   ```
## test cases result

<img width="1324" height="343" alt="Screenshot from 2025-08-15 02-04-41" src="https://github.com/user-attachments/assets/713a7de1-bddb-4918-93a8-3c1d087c0401" />



## Postman Collection

📥 **[Download Postman Collection](https://drive.google.com/file/d/1IlYumThjjYwqz4VBoLlfVNbB7hfjGY8i/view?usp=drive_link)**

## postman testcas


<img width="1791" height="767" alt="Screenshot from 2025-08-15 09-34-18" src="https://github.com/user-attachments/assets/664df848-ab8d-4fe2-8bef-5587207dfb42" />
e 


# Currency Conversion API

A REST API for fetching latest exchange rates, converting between currencies, and retrieving historical exchange rate data.

## Base URL

```
http://localhost:8080
```

---

## Endpoints

### 1. **Get Latest Exchange Rates**

Fetches the latest exchange rates based on a given base currency.

**Request**
```bash
curl -X GET "http://localhost:8080/latest?base=USD"
```

**Query Parameters**
| Name | Type   | Required | Description |
|------|--------|----------|-------------|
| base | String | Yes      | The base currency code (e.g., `USD`, `EUR`, `INR`). |

**Example Response**
```json
{
  "base": "USD",
  "date": "2025-08-15",
  "rates": {
    "INR": 83.21,
    "EUR": 0.91,
    "GBP": 0.78
  }
}
```

---

### 2. **Convert Currency (Latest Rate)**

Converts an amount from one currency to another using the latest available rates.

**Request**
```bash
curl -X GET "http://localhost:8080/convert?from=USD&to=INR&amount=100"
```

**Query Parameters**
| Name   | Type    | Required | Description |
|--------|---------|----------|-------------|
| from   | String  | Yes      | Source currency code. |
| to     | String  | Yes      | Target currency code. |
| amount | Number  | Yes      | Amount to convert. |

**Example Response**
```json
{
  "from": "USD",
  "to": "INR",
  "amount": 100,
  "converted": 8321,
  "rate": 83.21,
  "date": "2025-08-15"
}
```

---

### 3. **Convert Currency (Historical Rate)**

Converts an amount based on rates from a specific date.

**Request**
```bash
curl -X GET "http://localhost:8080/convert?from=USD&to=INR&amount=100&date=2024-08-14"
```
or
```bash
curl -X GET "http://localhost:8080/convert?from=BSD&to=INR&amount=100&date=2024-08-14"
```

**Query Parameters**
| Name   | Type    | Required | Description |
|--------|---------|----------|-------------|
| from   | String  | Yes      | Source currency code. |
| to     | String  | Yes      | Target currency code. |
| amount | Number  | Yes      | Amount to convert. |
| date   | String  | Yes      | Historical date in `YYYY-MM-DD` format. |

**Example Response**
```json
{
  "from": "USD",
  "to": "INR",
  "amount": 100,
  "converted": 8275,
  "rate": 82.75,
  "date": "2024-08-14"
}
```

---

### 4. **Get Historical Rate (Within 90 Days)**

Fetches historical exchange rates for a date within the last 90 days.

**Request**
```bash
curl -X GET "http://localhost:8080/historical?date=2025-08-01&source=USD&target=INR"
```

**Query Parameters**
| Name   | Type   | Required | Description |
|--------|--------|----------|-------------|
| date   | String | Yes      | Historical date in `YYYY-MM-DD`. |
| source | String | Yes      | Source currency code. |
| target | String | Yes      | Target currency code. |

**Example Response**
```json
{
  "date": "2025-08-01",
  "source": "USD",
  "target": "INR",
  "rate": 83.05
}
```

---

### 5. **Get Historical Rate (More than 90 Days)**

Fetches historical exchange rates for a date older than 90 days.

**Request**
```bash
curl -X GET "http://localhost:8080/historical?date=2025-03-01&source=USD&target=INR"
```

**Example Response**
```json
{
  "date": "2025-03-01",
  "source": "USD",
  "target": "INR",
  "rate": 82.10
}
```

---

## Error Handling

**Example Error Response**
```json
{
  "error": "date exceeds 90-day history limit."
}
```

Possible error:
- **400 Bad Request** – Missing or invalid parameters.


---


2. Use the `curl` commands above or import the [Postman Collection](postman_collection.json) to test.

---

## Project Structure

```
EXCHANGE-RATE-SERVICE/
├── .git/                      # Git version control
├── cmd/
│   └── server/
│       └── main.go             # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration handling
│   ├── delivery/
│   ├── handler/
│   │   ├── convert.go          # Convert currency handler
│   │   ├── handler_test.go     # Unit tests for handlers
│   │   ├── historical.go       # Historical data handler
│   │   ├── latest.go           # Latest rates handler
│   │   └── memetest.go         # Example/meme test handler
│   └── service/
│       └── latest.go           # Business logic for latest rates
├── .dockerignore               # Ignore files for Docker build
├── .env                        # Environment variables
├── .gitignore                  # Git ignore rules
├── docker-compose.yml          # Docker Compose configuration
├── Dockerfile                  # Docker build file
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── postman_collection.json     # Postman API collection
└── README.md                   # Project documentation
```



## Concurrency & Scaling Strategy

- **Efficient Concurrency Handling**  
  - API handlers are non-blocking where possible, ensuring high throughput.
    
- **Memcache for Caching**  
  - Used in-memory caching with **Memcache**, shared across instances.  
  - Automatically handles race conditions internally, so concurrent requests for the same key do not cause data corruption.  
  - Stores frequently requested exchange rates to reduce API latency and upstream API calls.  
  - TTL ensures stale data is automatically purged.  
  - Improves performance in high-traffic scenarios.  

- **Horizontal Scaling**  
  - Designed to run multiple instances behind a load balancer (e.g., Nginx, AWS ELB, Kubernetes Service).  
  - Stateless architecture ensures any instance can handle any request.  
  - Shared cache layer (Memcache) ensures consistent data across instances.  

- **Vertical Scaling**  
  - Optimized code paths and caching reduce CPU/memory usage.  
  - Can handle more requests per instance before scaling out.  

- **Memcache for Caching**  
  - Stores frequently requested exchange rates to reduce API latency and upstream API calls.  
  - TTL ensures stale data is automatically purged.  
  - Improves performance in high-traffic scenarios.  

- **used  `singleflight` in project**  
  - Prevents **cache stampede**: If multiple requests arrive for the same uncached data, only one request is sent to the upstream API.  
  - Other requests wait for the result, avoiding duplicate processing.  
  - Reduces load on both the application and the external exchange rate provider.  

- **High Availability Design**  
  - Cache + `singleflight` minimizes upstream dependency failures.  
  - Horizontal scaling allows rolling deployments without downtime.  

## Assumptions

- The exchangerate.host API is always available and responds correctly when `"success": true`.
- In-memory caching helps avoid hitting the API for every request, thus reducing response time.
- Since Memcached is being used, it automatically handles concurrency for cached data access.
- A background goroutine will periodically hit the API with base currency USD to fetch the updated exchange rates every hour while the server is running, keeping the cache fresh.

## This Backend service follow this architecture

<img width="594" height="274" alt="Screenshot from 2025-08-15 10-09-04" src="https://github.com/user-attachments/assets/ec22948e-4107-40c0-a9fb-140cfd4e4afe" />




