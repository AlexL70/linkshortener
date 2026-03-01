# URL Shortener Application Plan

This document outlines the idea and development plan for a URL shortener application.

## Initial idea

The initial idea for this project was to create a simple URL shortener application that registered users could use to shorten their long URLs. The application would allow users to log in with their Google, Facebook, or other third-party accounts (the exact list has not been finalized and is intended to cover as many users as possible) and manage their shortened URLs through a dashboard. The main requirements when choosing technologies were the following:

1. The initial version of the app should be hosted on a Linux server with 4GB of RAM on the cheapest DigitalOcean droplet.
2. The database should support SQL and be easy to scale in the future, including replication and sharding.
3. Both frontend and backend solutions should be lightning fast and easy to deploy on the same server. And neither of them should become a bottleneck for app performance, neither in the initial version nor in future versions as the app scales.
4. The app should support user registration and authentication through third-party providers (no passwords or sensitive information stored in the DB), the ability to add new third-party authentication credentials for the same user if necessary, without limiting the number of providers per user and accounts per provider, and the app should be able to handle the authentication flow for all supported providers without any issues. It should also implement CRUD for shortened URLs for authenticated users, with the option to allow users to set the shortcode themselves or have it generated automatically. It should also support an optional expiration date for the shortened URLs. The redirection should be lightning-fast, and the app should handle a large number of redirects without performance issues. Any performance bottlenecks should be easily identified and resolved by adding more server resources. But it is essential that the app uses resources as efficiently as possible. Simple statistics on redirects should be collected to set restrictions for free users in the future, if necessary, and to prevent DDoS attacks and other abuse of the service.
5. Code maintainability. The codebase should be clean, well-structured, strictly typed, and easy to understand. It is highly preferable to use an ORM to interact with the database and avoid writing raw SQL queries whenever possible. It is also desirable to use a code-first approach for database schema management and migrations. It is also desirable to have a way to automatically generate API calls and the data types necessary to make them on the frontend from the backend codebase, to avoid issues with data types and API endpoint mismatches. An OpenAPI specification for all API is a must-have and must be generated from the backend codebase. The codebase should also be well-documented and include comprehensive tests to ensure the app's reliability and maintainability.

# Development Plan

## 1. Backend (Go): API and Logic

- **Framework:** `Gin`
- **ORM:** `GORM` (github.com/go-gorm/gorm) with PostgreSQL driver (github.com/go-gorm/postgres)
  - Code-first approach: Define models as Go structs with GORM tags
  - Auto-migration support for development
  - Connection pooling and query optimization built-in
  - Support for future read/write splitting and sharding
- **Migrations:** `golang-migrate` (github.com/golang-migrate/migrate)
  - Version-controlled SQL migrations for production
  - Up/Down migration support
  - CLI tool for migration management
- **OpenAPI Specification:** `Swag` (github.com/swaggo/swag) with `gin-swagger` (github.com/swaggo/gin-swagger)
  - Auto-generate OpenAPI 3.0 spec from Go code comments
  - Interactive API documentation (Swagger UI)
  - Single source of truth for API contracts
- **Authentication & Session Management:**
  - `go-oidc` (github.com/coreos/go-oidc) for OIDC authentication
  - JWT tokens (`github.com/golang-jwt/jwt`) for session management
  - Supported OAuth2/OIDC Providers (MVP):
    - Google (via google.com)
    - Microsoft (via login.microsoftonline.com)
    - Facebook (via graph.facebook.com)
  - Future extensibility: Architecture supports adding GitHub, Apple, LinkedIn, etc.
- **URL Shortening Logic:** Base62 encoding, URL validation
- **API Endpoints:**
  - `/r/{shortcode}`: **PUBLIC** - Redirect to the long URL (no authentication required)
  - `/shorten`: Create a short URL record (JWT protected)
  - `/user/urls`: List user's shortened URLs (JWT protected)
  - `/user/urls/{id}`: CRUD for specific user URLs (JWT protected)
  - `/auth/login/{provider}`: **PUBLIC** - Initiate OAuth2/OIDC flow with specified provider
  - `/auth/callback`: **PUBLIC** - OAuth2/OIDC callback endpoint to complete authentication
  - `/auth/logout`: Invalidate JWT token (JWT protected)

**Endpoint Protection Summary:**

- **Public (No JWT Required):** `/r/{shortcode}`, `/auth/login/{provider}`, `/auth/callback`
- **Protected (JWT Required):** All other endpoints (`/shorten`, `/user/urls`, `/user/urls/*`, `/auth/logout`, etc.)

### Authentication Flow & JWT Management

**Session Management Strategy:**

- All authenticated requests use JWT tokens in `Authorization: Bearer <token>` header
- JWT tokens include `user_id`, `subject`, `issued_at`, `expiration`
- No server-side session storage (stateless, scalable)
- Refresh tokens handled client-side with token rotation

**MVP Provider Registration Flow:**

1. Frontend user selects provider (Google, Microsoft, or Facebook)
2. Frontend redirects to backend `/auth/login/{provider}`
3. Backend initiates OIDC flow using `go-oidc`
4. Provider redirects user to login/consent screen
5. Provider redirects back to `/auth/callback` with authorization code
6. Backend exchanges code for ID token and user info
7. Backend checks if user exists in `UserProviders` table:
   - **Existing user:** Return JWT token
   - **New user:** Create user in `Users` table with username from provider, create entry in `UserProviders`, return JWT token
8. Frontend stores JWT token and redirects to dashboard

**Future Extension:**
New providers (GitHub, Apple, LinkedIn, etc.) can be added by:

- Adding new provider config to backend
- Adding new login button to frontend
- No database schema changes required (already supports unlimited providers per user)

## 2. Database (PostgreSQL): Schema

- **Users Table:**
  - `id` (SERIAL PRIMARY KEY)
  - `user_name` (VARCHAR UNIQUE NOT NULL) - Initially from first provider, user can change
  - `created_at` (TIMESTAMP WITH TIME ZONE)
  - `updated_at` (TIMESTAMP WITH TIME ZONE)
- **UserProviders Table:**
  - `id` (SERIAL PRIMARY KEY)
  - `user_id` (INTEGER REFERENCES Users(id) ON DELETE CASCADE)
  - `provider` (VARCHAR NOT NULL) - e.g., "google", "facebook", "github"
  - `provider_user_id` (VARCHAR NOT NULL) - User ID from the provider
  - `provider_email` (VARCHAR) - Optional, for display purposes
  - `created_at` (TIMESTAMP WITH TIME ZONE)
  - **UNIQUE** constraint on (`provider`, `provider_user_id`)
  - **INDEX** on `user_id` for fast user lookup
  - **NOTE:** Allows multiple providers per user and multiple accounts per provider

- **ShortenedUrls Table:**
  - `id` (SERIAL PRIMARY KEY)
  - `user_id` (INTEGER REFERENCES Users(id) ON DELETE CASCADE)
  - `shortcode` (VARCHAR UNIQUE NOT NULL)
  - `long_url` (TEXT NOT NULL)
  - `expires_at` (TIMESTAMP WITH TIME ZONE) - **NULLABLE** for optional expiration
  - `created_at` (TIMESTAMP WITH TIME ZONE)
  - `updated_at` (TIMESTAMP WITH TIME ZONE)
  - **INDEX** on `shortcode` for fast redirect lookups
  - **INDEX** on `user_id` for user's URL listing

- **UrlClicks Table:**
  - `id` (SERIAL PRIMARY KEY)
  - `url_id` (INTEGER REFERENCES ShortenedUrls(id) ON DELETE CASCADE)
  - `clicked_at` (TIMESTAMP WITH TIME ZONE)
  - `ip_address` (VARCHAR) - For rate limiting and abuse detection
  - `user_agent` (TEXT) - Optional, for basic analytics
  - `referer` (TEXT) - Optional, for traffic source tracking
  - **INDEX** on `url_id` for per-URL statistics
  - **INDEX** on `clicked_at` for time-based queries
  - **INDEX** on `ip_address` for rate limiting and abuse prevention
  - **NOTE:** Consider partitioning by date for long-term scalability

## 3. Frontend (Vue.js): User Interface

- **Framework:** Vue.js
- **UI Library:** Vue Material, Vuetify
- **Type-Safe API Client:** `openapi-typescript-codegen`
  - Auto-generate TypeScript types from OpenAPI spec
  - Auto-generate typed API client with all HTTP methods
  - Ensures type safety between frontend and backend
- **Authentication:**
  - Google Sign-In SDK (via @react-oauth/google or @socialpure/vue-google-login)
  - Microsoft Authentication Library (MSAL) for Vue
  - Facebook SDK for JavaScript
  - JWT token storage (localStorage or secure cookie)
  - Middleware for route protection
- **State Management:** Pinia or Vuex for JWT token and user state
- **Components:**
  - Login/Signup Page
  - URL Shortening Form
  - Dashboard

## 4. Resource Efficiency (4GB RAM constraint)

### Environment Variables Configuration

```bash
# Database Connection Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=3600        # seconds (1 hour)
DB_CONN_MAX_IDLE_TIME=600        # seconds (10 minutes)

# HTTP Server
REQUEST_TIMEOUT_SECONDS=30
MAX_CONCURRENT_REQUESTS=100

# Batch Processing
CLICK_BATCH_SIZE=1000
CLICK_BATCH_TIMEOUT_SECONDS=5

# Pagination
DEFAULT_PAGE_SIZE=20
```

### GORM Configuration

```go
// Connection pool tuned for 4GB RAM
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    // Disable prepared statement caching to save memory
    PrepareStmt: false,
    // Use single connection for read operations where possible
    NowFunc: func() time.Time { return time.Now().UTC() },
})

// Set connection pool limits from environment
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(os.Getenv("DB_MAX_OPEN_CONNS"))        // default: 25
sqlDB.SetMaxIdleConns(os.Getenv("DB_MAX_IDLE_CONNS"))        // default: 5
sqlDB.SetConnMaxLifetime(time.Duration(os.Getenv("DB_CONN_MAX_LIFETIME")) * time.Second)   // default: 1 hour
sqlDB.SetConnMaxIdleTime(time.Duration(os.Getenv("DB_CONN_MAX_IDLE_TIME")) * time.Second)  // default: 10 min
```

### Gin Configuration

```go
// Production mode to disable debug logging overhead
gin.SetMode(gin.ReleaseMode)

// Create Gin engine with pooled writers
engine := gin.New()

// Add middleware for request timeouts
requestTimeout := os.Getenv("REQUEST_TIMEOUT_SECONDS")  // default: 30
engine.Use(timeout.Middleware(time.Duration(requestTimeout) * time.Second))

// Limit concurrent requests to prevent memory spikes
maxConcurrent := os.Getenv("MAX_CONCURRENT_REQUESTS")  // default: 100
engine.Use(limiter.Middleware(maxConcurrent))
```

### PostgreSQL Configuration (1-2GB RAM)

Add to `postgresql.conf`:

```ini
# Memory Settings (total ~800MB for 4GB system)
shared_buffers = 200MB              # 25% of available RAM
effective_cache_size = 600MB        # 75% of available RAM
work_mem = 4MB                      # Per operation memory (low for many concurrent ops)
maintenance_work_mem = 50MB         # For VACUUM, CREATE INDEX

# Connections
max_connections = 50                # Conservative for 4GB droplet
superuser_reserved_connections = 3

# WAL Settings (for reliability without heavy I/O)
wal_buffers = 16MB
checkpoint_timeout = 15min
max_wal_size = 2GB

# Query Planning
random_page_cost = 1.1              # Assume SSD storage (DigitalOcean default)
effective_io_concurrency = 200      # For modern storage

# Logging (minimal overhead)
log_min_duration_statement = 1000   # Log queries > 1 second
log_connections = off               # Disable connection logging

# Temp Files
temp_file_limit = 1GB               # Prevent runaway query temp files

# Autovacuum (aggressive for small datasets)
autovacuum = on
autovacuum_min_maint_tuples = 50
```

### Application-Level Optimizations

- **Batch URL Clicks:** Write clicks to in-memory buffer, flush every `$CLICK_BATCH_TIMEOUT_SECONDS` seconds or `$CLICK_BATCH_SIZE` records (defaults: 5 seconds or 1000 records)
- **Pagination:** All list endpoints default to page size `$DEFAULT_PAGE_SIZE` (default: 20)
- **Query Optimization:** Always use specific columns in SELECT, never `SELECT *`
- **Connection Pooling:** GORM handles this; monitor with `/debug/pprof`
- **Goroutine Limits:** Middleware caps concurrent requests at `$MAX_CONCURRENT_REQUESTS` (default: 100)
- **Request Timeouts:** All handlers have `$REQUEST_TIMEOUT_SECONDS` timeout to prevent hanging connections (default: 30 seconds)

## 5. Security & Rate Limiting Strategy

### Environment Variables Configuration

```bash
# Monthly Quota Settings
MONTHLY_QUOTA_CLICKS=1000

# Per-Minute Rate Limiting (DDoS Prevention)
RATE_LIMIT_PER_MINUTE=100

# Suspension & Timeout
SUSPENSION_TIMEOUT_MINUTES=60

# Input Validation
MAX_URL_LENGTH=2048
MIN_SHORTCODE_LENGTH=6
MAX_SHORTCODE_LENGTH=6
```

### Monthly Quota (MVP - Prevents Resource Abuse)

- **Quota:** `$MONTHLY_QUOTA_CLICKS` (default: 1000 clicks per month per user)
- **Tracking:** Store in `UserQuotas` table:
  - `user_id` (FOREIGN KEY)
  - `period_start` (timestamp of month start)
  - `clicks_used` (counter)
  - `reset_at` (TIMESTAMP when quota resets)
- **Enforcement:**
  - Check `clicks_used < $MONTHLY_QUOTA_CLICKS` before processing redirect
  - On quota exceeded: Return 429 HTTP status with message to upgrade
  - Reset automatically on first of next month (check `reset_at` timestamp)
- **Future Upgrade Path:** Paid tiers can have higher limits (configurable per tier)

### Per-Minute Rate Limiting (DDoS Prevention)

- **Limit:** `$RATE_LIMIT_PER_MINUTE` (default: 100 clicks per minute per user)
- **Tracking:** In-memory counter per user (reset every minute)
  - Use `sync.Map` with cleanup goroutine or simple map with lock
  - Alternative: Use `UrlClicks` table with time window query (heavier but persistent)
- **Detection & Response:**
  - Track minute-by-minute activity in `UserAnomalies` table:
    - `user_id`
    - `detected_at` (TIMESTAMP)
    - `clicks_in_minute` (count)
    - `anomaly_type` ("rate_limit_exceeded")
    - `status` ("flagged", "suspended")
  - If `$RATE_LIMIT_PER_MINUTE` clicks/minute exceeded:
    - Log anomaly to `UserAnomalies` table
    - Set user status to `suspended`
    - Add `suspended_until` (TIMESTAMP) - calculated as `now() + $SUSPENSION_TIMEOUT_MINUTES`
- **Suspension Mechanism:**
  - Middleware checks user's `suspended_until` timestamp before processing
  - If suspended: Return 429 "Account temporarily suspended due to suspicious activity"
  - Automatic unsuspend after timeout expires
  - Manual override: Admin can unsuspend earlier
- **Monitoring:** Alert on any anomalies (email to admin)

### Input Validation

- **URL Validation:**
  - Verify HTTP/HTTPS URLs only
  - Reject localhost, private IP ranges (prevent SSRF)
  - Max URL length: `$MAX_URL_LENGTH` (default: 2048 characters)
- **Shortcode Validation:**
  - If user-provided: alphanumeric + hyphens, fixed length of `$MAX_SHORTCODE_LENGTH` characters (default: 6)
  - Check against reserved words (admin, api, www, etc.)
  - If auto-generated: use Base62 encoding, always exactly 6 characters, ensure uniqueness
  - **URL Format:** `domain.com/r/{6-char-shortcode}` = max 20 characters (e.g., `example.com/r/abc123`)
- **User Input:** Trim whitespace, validate UTF-8 encoding

## 6. Testing Strategy

### Backend (Go): Unit Tests - 95% Coverage Minimum

**Testing Framework & Tools:**

- **Test Framework:** Built-in `testing` package (Go standard library)
- **Assertions:** `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`
- **Mocking:** `github.com/golang/mock` (mockgen) for interfaces
- **Coverage Tracking:** Built-in `go test -cover` and integration with CI/CD

**Coverage Requirements:**

- **Minimum:** 95% code coverage for all packages
- **Exclusions:** Only main() function and vendor packages
- **CI/CD Integration:** Fail build if coverage drops below 95%

**Testing Structure:**

```go
// Example unit test file: handlers/urls_test.go
package handlers

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestShortenURL_Success(t *testing.T) {
    // Arrange: Setup test data and mocks
    mockDB := NewMockDB()
    handler := NewURLHandler(mockDB)

    // Act: Execute function
    result, err := handler.Shorten(ctx, &ShortenRequest{
        LongURL: "https://example.com/very/long/url",
    })

    // Assert: Verify results
    require.NoError(t, err)
    assert.Equal(t, 6, len(result.Shortcode))
}

func TestShortenURL_InvalidURL(t *testing.T) {
    // Test edge cases: empty URL, invalid protocol, etc.
}

func TestRateLimit_ExceededQuota(t *testing.T) {
    // Test rate limiting logic
}
```

**Test Coverage Areas:**

- **Authentication:** JWT validation, provider OIDC flows, token refresh
- **URL Operations:** Creation, retrieval, deletion, expiration logic
- **Rate Limiting:** Monthly quota enforcement, per-minute DDoS detection
- **Database Layer:** GORM interactions, error handling
- **Input Validation:** URL/shortcode validation, SSRF prevention
- **Error Handling:** All HTTP error responses

**Command to Check Coverage:**

```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out  # Generate HTML report
```

---

### Frontend (Vue.js): Unit Tests - 80% Coverage Minimum

**Testing Framework & Tools:**

- **Test Runner:** `Vitest` (github.com/vitest-dev/vitest)
- **Component Testing:** `@vue/test-utils` (Vue 3 compatible)
- **Assertions:** Vitest built-in expect API
- **Mocking:** `vi.mock()` and `vi.spyOn()`
- **Coverage:** `@vicover/coverage` plugin built-in with Vitest

**Coverage Requirements:**

- **Minimum:** 80% code coverage for component logic
- **Exclusions:** CSS/styling, third-party SDK integrations, complex animations
- **Focus Areas:** Business logic, state management, API integration

**Testing Structure:**

```javascript
// Example unit test: LoginForm.spec.js
import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import LoginForm from "@/components/LoginForm.vue";
import { useAuthStore } from "@/stores/auth";

describe("LoginForm.vue", () => {
  let wrapper;
  let authStore;

  beforeEach(() => {
    wrapper = mount(LoginForm);
    authStore = useAuthStore();
  });

  it("renders MVP provider buttons correctly", () => {
    expect(wrapper.text()).toContain("Sign in with Google");
    expect(wrapper.text()).toContain("Sign in with Microsoft");
    expect(wrapper.text()).toContain("Sign in with Facebook");
  });

  it("handles login click for Google provider", async () => {
    const googleBtn = wrapper.find('[data-test="login-google"]');
    await googleBtn.trigger("click");
    expect(authStore.initiateLogin).toHaveBeenCalledWith("google");
  });

  it("displays error message on failed login", async () => {
    authStore.loginError = "Authentication failed";
    await wrapper.vm.$nextTick();
    expect(wrapper.text()).toContain("Authentication failed");
  });
});

describe("URLShortenerForm.vue", () => {
  it("validates URL before submission", async () => {
    const form = mount(URLShortenerForm);
    await form.find("input").setValue("invalid-url");
    await form.find("form").trigger("submit");
    expect(form.vm.errors).toContain("Invalid URL");
  });

  it("submits valid URL and displays shortened link", async () => {
    const form = mount(URLShortenerForm);
    await form.find("input").setValue("https://example.com/long");
    await form.find("form").trigger("submit");
    expect(form.vm.shortenedURL).toBeTruthy();
  });
});

describe("UserStore", () => {
  it("stores JWT token after login", () => {
    const store = useAuthStore();
    store.setToken("eyJhbGc...");
    expect(store.token).toBe("eyJhbGc...");
  });

  it("clears token on logout", () => {
    const store = useAuthStore();
    store.logout();
    expect(store.token).toBeNull();
  });
});
```

**Test Coverage Areas:**

- **Authentication Components:** Login form, provider button rendering
- **URL Shortening:** Form validation, submission, error display
- **State Management:** Token storage, user state, logout flow
- **API Integration:** Mock API calls, success/error handling
- **Route Guards:** Protected route access control

**Commands to Run Tests:**

```bash
npm run test                    # Run tests in watch mode
npm run test:ui                # Interactive UI mode
npm run coverage               # Generate coverage report
npm run test:coverage:report   # HTML coverage report
```

**Vitest Configuration (vitest.config.js):**

```javascript
import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: "jsdom",
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      exclude: ["node_modules/", "dist/", "**/*.d.ts", "**/types/**"],
      lines: 80,
      functions: 80,
      branches: 80,
      statements: 80,
    },
  },
});
```

## 7. Deployment Strategy (MVP - Single Docker Container)

### Architecture: All-in-One Container

**Single Docker image containing:**

- PostgreSQL database
- Go backend API (compiled binary)
- Vue.js frontend (built/dist folder)
- Nginx reverse proxy (frontend + API routing)
- Database migrations (golang-migrate)

**Advantages for MVP:**

- Simple, single deployment unit
- Everything runs on one 4GB DigitalOcean droplet
- Easy to version the entire stack
- Straightforward rollback (just restart container)

### Build Process & Dockerfile

**Multi-stage Dockerfile** (optimized for 4GB RAM):

```dockerfile
# Stage 1: Build Go backend
FROM golang:1.21-alpine AS backend-builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

# Stage 2: Build Vue.js frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /build
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 3: Runtime image
FROM postgres:16-alpine
RUN apk add --no-cache nginx curl

# Copy PostgreSQL config (tuned for 4GB RAM)
COPY deployment/postgresql.conf /etc/postgresql/postgresql.conf

# Copy Go API binary
COPY --from=backend-builder /build/api /usr/local/bin/api

# Copy compiled frontend
COPY --from=frontend-builder /build/dist /var/www/html

# Copy Nginx configuration
COPY deployment/nginx.conf /etc/nginx/nginx.conf

# Copy migrations
COPY migrations/ /migrations/

# Copy startup script
COPY deployment/start.sh /start.sh
RUN chmod +x /start.sh

# Expose ports
EXPOSE 80 443 5432

ENTRYPOINT ["/start.sh"]
```

### Docker Compose (Local Development Reference)

```yaml
# docker-compose.yml
version: "3.9"

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "80:80"
      - "5432:5432"
    environment:
      - DB_HOST=localhost
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=linkshortener
      - MONTHLY_QUOTA_CLICKS=1000
      - RATE_LIMIT_PER_MINUTE=100
      - SUSPENSION_TIMEOUT_MINUTES=60
      - JWT_SECRET=dev-secret-change-in-production
      - OAUTH_GOOGLE_CLIENT_ID=${OAUTH_GOOGLE_CLIENT_ID}
      - OAUTH_GOOGLE_CLIENT_SECRET=${OAUTH_GOOGLE_CLIENT_SECRET}
      - OAUTH_MICROSOFT_CLIENT_ID=${OAUTH_MICROSOFT_CLIENT_ID}
      - OAUTH_MICROSOFT_CLIENT_SECRET=${OAUTH_MICROSOFT_CLIENT_SECRET}
      - OAUTH_FACEBOOK_CLIENT_ID=${OAUTH_FACEBOOK_CLIENT_ID}
      - OAUTH_FACEBOOK_CLIENT_SECRET=${OAUTH_FACEBOOK_CLIENT_SECRET}
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### Startup Script (start.sh)

```bash
#!/bin/bash
set -e

# Start PostgreSQL
echo "Starting PostgreSQL..."
postgres -D /var/lib/postgresql/data &
sleep 5

# Wait for PostgreSQL to be ready
until psql -U postgres -c '\q'; do
  echo "Postgres is unavailable, sleeping..."
  sleep 2
done

# Create database if not exists
createdb -U postgres linkshortener || true

# Run migrations
/usr/local/bin/migrate -path /migrations -database \
  "postgres://postgres:${DB_PASSWORD}@localhost:5432/linkshortener?sslmode=disable" up

# Start Nginx (reverse proxy)
echo "Starting Nginx..."
nginx -g 'daemon off;' &

# Start Go API
echo "Starting Go API..."
API_HOST=0.0.0.0 API_PORT=8080 /usr/local/bin/api

wait
```

### Nginx Configuration (nginx.conf)

```nginx
events {
    worker_connections 100;
}

http {
    # Serve frontend static files
    server {
        listen 80 default_server;
        server_name _;

        # Frontend static assets
        location / {
            root /var/www/html;
            try_files $uri $uri/ /index.html;
            expires 1d;
            add_header Cache-Control "public, immutable";
        }

        # API proxy
        location /api/ {
            proxy_pass http://localhost:8080/api/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_read_timeout 30s;
        }

        # Short redirect endpoint (public, no rate limiting by auth)
        location /r/ {
            proxy_pass http://localhost:8080/r/;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_read_timeout 10s;
        }

        # Swagger UI (if enabled)
        location /swagger/ {
            proxy_pass http://localhost:8080/swagger/;
        }
    }
}
```

### Environment Configuration

**Production Environment Variables (.env):**

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<secure-password>
DB_NAME=linkshortener

# API
API_HOST=0.0.0.0
API_PORT=8080
JWT_SECRET=<generate-secure-secret>

# Rate Limiting & Quotas
MONTHLY_QUOTA_CLICKS=1000
RATE_LIMIT_PER_MINUTE=100
SUSPENSION_TIMEOUT_MINUTES=60

# OAuth2/OIDC Credentials (from provider dashboards)
OAUTH_GOOGLE_CLIENT_ID=<your-google-client-id>
OAUTH_GOOGLE_CLIENT_SECRET=<your-google-secret>
OAUTH_MICROSOFT_CLIENT_ID=<your-microsoft-client-id>
OAUTH_MICROSOFT_CLIENT_SECRET=<your-microsoft-secret>
OAUTH_FACEBOOK_CLIENT_ID=<your-facebook-app-id>
OAUTH_FACEBOOK_CLIENT_SECRET=<your-facebook-secret>

# Resource Limits
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
MAX_CONCURRENT_REQUESTS=100
REQUEST_TIMEOUT_SECONDS=30
```

### Deployment Steps on DigitalOcean Droplet

1. **Create DigitalOcean Droplet:**
   - Select Ubuntu 22.04 LTS, 4GB RAM, SSD storage
   - Configure firewall: Allow ports 22 (SSH), 80 (HTTP), 443 (HTTPS)

2. **Initial Server Setup:**

   ```bash
   ssh root@<droplet-ip>
   apt update && apt upgrade -y
   apt install -y docker.io docker-compose
   usermod -aG docker $USER
   ```

3. **Deploy Application:**

   ```bash
   # Clone repository
   git clone <repo-url> /opt/linkshortener
   cd /opt/linkshortener

   # Create .env file with production secrets
   nano .env

   # Build image
   docker build -t linkshortener:latest .

   # Run container
   docker run -d \
     --name linkshortener \
     --restart always \
     --env-file .env \
     -p 80:80 \
     -p 443:443 \
     -v linkshortener_db:/var/lib/postgresql/data \
     linkshortener:latest
   ```

4. **Setup SSL (Let's Encrypt):**

   ```bash
   # Install Certbot
   apt install -y certbot python3-certbot-nginx

   # Generate certificate
   certbot certonly --standalone -d <your-domain>

   # Update Nginx config to use SSL
   # (Update nginx.conf with SSL certificates)
   ```

5. **Monitor & Maintain:**

   ```bash
   # View logs
   docker logs -f linkshortener

   # Check resource usage
   docker stats linkshortener

   # Backup database
   docker exec linkshortener pg_dump -U postgres linkshortener > backup.sql
   ```

### 8. Logging & Monitoring Strategy (Lightweight for 4GB RAM)

**Philosophy:** Minimal overhead, maximum observability within memory constraints.

#### Backend Logging

**Logging Framework:**

- **Library:** Go 1.21+ standard library `slog` (zero external dependencies)
- **Output:** stdout only (Docker captures and manages logs)
- **Format:** JSON for easy parsing and indexing
- **Levels:** ERROR, WARN, INFO (no DEBUG in production)

**Log Targets:**

- **errors:** All HTTP errors, database errors, authentication failures
- **warnings:** Rate limit violations, quota near limit (90%+), suspicious patterns
- **info:** Request count by endpoint, slow queries (>1s), anomalies detected

```go
// Example logging
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

logger.Error("database error", "query", "SELECT...", "error", err)
logger.Warn("rate limit exceeded", "user_id", 123, "ip", "192.168.1.1")
logger.Info("url shortened", "user_id", 123, "shortcode", "abc123")
```

**View Logs:**

```bash
docker logs -f linkshortener              # Real-time logs with auto-formatting
docker logs linkshortener | tail -100     # Last 100 entries
docker logs linkshortener > app.log       # Save to file for analysis
```

#### Metrics Collection (In-Memory)

**Lightweight Metrics Struct:**

```go
type Metrics struct {
    TotalRequests       int64
    TotalErrors         int64
    ActiveConnections   int
    ActiveGoroutines    int
    RateLimitViolations int64
    AnomaliesDetected   int64
    AvgResponseTimeMs   float64
    Timestamp           time.Time
}
```

**Exposed via `/api/metrics` Endpoint (Plain text Prometheus format):**

```
# HELP requests_total Total HTTP requests processed
# TYPE requests_total counter
requests_total 15234

# HELP errors_total Total HTTP errors
# TYPE errors_total counter
errors_total 42

# HELP active_connections Current active database connections
# TYPE active_connections gauge
active_connections 12

# HELP rate_limit_violations_total Rate limit violations detected
# TYPE rate_limit_violations_total counter
rate_limit_violations_total 3

# HELP anomalies_detected_total Anomalies detected (suspicious activity)
# TYPE anomalies_detected_total counter
anomalies_detected_total 1

# HELP response_time_avg_ms Average response time in milliseconds
# TYPE response_time_avg_ms gauge
response_time_avg_ms 45.3
```

**Monitoring via Cron Job:**

```bash
#!/bin/bash
# /opt/linkshortener/monitoring.sh - Run every 5 minutes
# 0 */5 * * * /opt/linkshortener/monitoring.sh

METRICS=$(curl -s http://localhost:8080/api/metrics)

# Check response time threshold (>500ms is concerning on 4GB)
RESPONSE_TIME=$(echo "$METRICS" | grep response_time_avg_ms | awk '{print $2}')
if (( $(echo "$RESPONSE_TIME > 500" | bc -l) )); then
    echo "ALERT: High response time: ${RESPONSE_TIME}ms" | mail -s "LinkShortener Alert" admin@example.com
fi

# Check error rate (>1% is concerning for MVP)
ERRORS=$(echo "$METRICS" | grep "^errors_total" | awk '{print $2}')
TOTAL=$(echo "$METRICS" | grep "^requests_total" | awk '{print $2}')
ERROR_RATE=$(echo "scale=4; $ERRORS / $TOTAL * 100" | bc)
if (( $(echo "$ERROR_RATE > 1" | bc -l) )); then
    echo "ALERT: Error rate: ${ERROR_RATE}%" | mail -s "LinkShortener Alert" admin@example.com
fi

# Check active connections (>40 is concerning on single container)
CONNECTIONS=$(echo "$METRICS" | grep "^active_connections" | awk '{print $2}')
if (( CONNECTIONS > 40 )); then
    echo "ALERT: High connection count: $CONNECTIONS" | mail -s "LinkShortener Alert" admin@example.com
fi

# Log metrics for analysis
echo "$(date '+%Y-%m-%d %H:%M:%S'): $METRICS" >> /opt/linkshortener/metrics.log
```

#### Alternative: DigitalOcean Native Monitoring

Use DigitalOcean's built-in droplet monitoring dashboard (no agent needed):

- **CPU Usage:** Alert if >70% sustained
- **Memory Usage:** Alert if >80% sustained
- **Disk I/O:** Monitor for spikes
- **Bandwidth:** Track traffic patterns

#### Alerts Configuration

**Email Alerts (via `ssmtp`):**

```bash
# Install
apt install -y ssmtp

# Configure /etc/ssmtp/ssmtp.conf
root=admin@example.com
mailhub=smtp.gmail.com:587
AuthUser=your-email@gmail.com
AuthPass=your-app-password
UseSTARTTLS=YES
```

**Thresholds for MVP:**

- ERROR logs: Immediate alert
- Response time > 500ms for 5+ consecutive requests: Alert
- Error rate > 1%: Alert
- Anomalies detected: Immediate alert
- Memory usage > 3.5GB: Alert
- Disk usage > 80%: Alert

#### Memory Footprint

- **slog logging:** <1MB
- **In-memory metrics:** ~2-5MB (minimal data retention)
- **Cron monitoring:** Negligible
- **Total overhead:** <10MB (well within 4GB budget)

#### Future Upgrades

When app grows beyond MVP:

- Add Prometheus + Grafana (separate container)
- Send metrics to managed service (DataDog, New Relic, or DigitalOcean)
- Implement distributed tracing (Jaeger)
- Add application performance monitoring (APM)

### Future Scaling Considerations

When ready to move beyond MVP:

- **Separate Containers:** Run PostgreSQL, API, and frontend in separate containers with Docker Compose or Kubernetes
- **Database Replication:** Run read replicas in separate containers
- **Horizontal Scaling:** API instances behind load balancer
- **CDN:** Serve frontend from CDN instead of docker container
- **Container Registry:** Push images to Docker Hub or private registry for version management

## Key Considerations & Next Steps:

- **Error Handling:** Comprehensive error handling with user-friendly messages.
- **Scalability Testing:** Load testing, performance optimization.
- **Database Indexing:** Indexes on `user_id` and `shortcode`.
- **Monitoring:** Use Go's pprof to track memory, goroutines, and connection pool health.
- **Anomaly Alerts:** Email/Slack notifications for suspicious activity patterns.
- **SSL/TLS:** Configure HTTPS with Let's Encrypt for production security.
