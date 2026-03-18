# Deployment Guide

## Prerequisites & Restrictions

- **Target OS:** Ubuntu 24.04 (or any modern Linux with systemd)
- **Target architecture:** `linux/amd64`
- **Required server software (pre-installed, not managed by this project):**
  - Caddy v2.8.4+ (with ports 80 and 443 open to the internet for automatic TLS)
  - PostgreSQL 18.3+
- **Build machine:** your local Linux/macOS development machine (Go toolchain + Node.js required)
- **Access:** SSH access to the target server with `sudo` privileges
- **DNS:** a domain name pointed at the server's IP address (required for Caddy's automatic Let's Encrypt certificate)
- **Auto-migration is disabled in production.** The schema must be applied manually before the first start, and individual migration SQL files must be applied manually on each subsequent schema change. See steps 4 and 7 below.

---

## Phase 1 — One-time server setup

Run these steps once when provisioning a new server.

### Step 1 — Create the PostgreSQL database and user

SSH into the server and run:

```bash
sudo -u postgres psql -c "CREATE USER linkshortener WITH PASSWORD 'your-strong-password';"
sudo -u postgres psql -c "CREATE DATABASE linkshortener OWNER linkshortener;"
```

### Step 2 — Run the setup script

From your **local machine**, copy the `deploy/` directory to the server and execute the setup script:

```bash
scp -r deploy/* user@your-server:/tmp/ls-setup/
ssh user@your-server "cd /tmp/ls-setup && sudo bash server-setup.sh"
```

The script (idempotent — safe to re-run):

- Creates the `linkshortener` system user
- Creates `/opt/linkshortener/frontend/` and `/etc/linkshortener/`
- Copies `deploy/environment.template` to `/etc/linkshortener/environment` (**only if the file does not already exist**, so re-runs preserve your secrets)
- Installs and enables the `linkshortener.service` systemd unit

### Step 3 — Fill in the environment file

SSH into the server:

```bash
ssh user@your-server "sudo nano /etc/linkshortener/environment"
```

Replace every `CHANGE_ME` placeholder. Generate cryptographically strong secrets (at least 32 bytes) with:

```bash
openssl rand -hex 32
```

Required values to set:

- `DATABASE_URL` — e.g. `postgres://linkshortener:your-strong-password@localhost:5432/linkshortener`
- `JWT_SECRET`
- `SESSION_SECRET`
- `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`
- `MICROSOFT_CLIENT_ID` and `MICROSOFT_CLIENT_SECRET`
- `FACEBOOK_CLIENT_ID` and `FACEBOOK_CLIENT_SECRET`
- `APP_BASE_URL` — e.g. `https://yourdomain.com`
- `SUPER_ADMIN_EMAIL`

### Step 4 — Apply the database schema

From your **local machine**:

```bash
cat sql/schema.sql | ssh user@your-server "sudo -u postgres psql linkshortener"
```

### Step 5 — Install the Caddyfile

Edit `deploy/Caddyfile` on your local machine: replace `yourdomain.com` with your actual domain, then copy it to the server:

```bash
scp deploy/Caddyfile user@your-server:/tmp/Caddyfile
ssh user@your-server "sudo cp /tmp/Caddyfile /etc/caddy/Caddyfile && sudo systemctl reload caddy"
```

### Step 6 — First deployment

```bash
make deploy SSH_HOST=user@your-server SSH_KEY=<your_key> APP_BASE_URL=https://yourdomain.com
```

`SSH_KEY` is optional if your SSH config already specifies the key for that host.

This builds the Go binary and the Vue frontend locally and rsyncs them to the server, then restarts the service. See [Phase 2](#phase-2--repeatable-deployment) for details.

### Step 7 — Verify

```bash
ssh user@your-server "sudo systemctl status linkshortener"
curl -I https://yourdomain.com                   # expect HTTP 200, TLS cert present
curl -I https://yourdomain.com/r/nonexistent     # expect JSON 404 from the Go backend
```

---

## Phase 2 — Repeatable deployment

Every subsequent deployment after Phase 1 follows this sequence:

### Step 1 — Bump the version

**Every deployment must start with a version bump.** The version is stored in the `version` file at the repository root and is baked into both the backend binary and the frontend build at compile time.

Use the `bump` Makefile target with one of three arguments:

| Argument     | Effect                                                   | Example: current `1.2.3` → new |
| ------------ | -------------------------------------------------------- | ------------------------------ |
| `BUMP=patch` | Increments the third number; resets nothing              | `1.2.4`                        |
| `BUMP=minor` | Increments the second number; resets patch to 0          | `1.3.0`                        |
| `BUMP=major` | Increments the first number; resets minor and patch to 0 | `2.0.0`                        |

```bash
# Choose one:
make bump BUMP=patch    # bug fixes, small tweaks
make bump BUMP=minor    # new features, backwards-compatible changes
make bump BUMP=major    # breaking changes, major milestones
```

What `make bump` does automatically:

1. Reads the current version from the `version` file
2. Increments the requested component and writes the new version back
3. Stages **all** pending local changes (`git add -A`) — including your code changes, not just the version file
4. Creates a commit with the message `Version bumped to X.Y.Z`
5. Creates a lightweight git tag `vX.Y.Z` pointing at that commit

After bumping, push the commit and tag to the remote when ready:

```bash
git push && git push --tags
```

### Step 2 — Apply schema migrations (if any)

When a release includes new database migrations, apply the relevant SQL files **before** running `make deploy`:

```bash
cat sql/00N_migration_name.sql | ssh user@your-server "sudo -u postgres psql linkshortener"
```

Migration files are named with a sequential numeric prefix (e.g. `002_add_clicks_table.sql`). Apply only the files that are new since the last deployment, in order.

### Step 3 — Deploy

```bash
make deploy SSH_HOST=user@your-server SSH_KEY=~/.ssh/<your_key> APP_BASE_URL=https://yourdomain.com
```

This runs the following steps automatically:

1. Compiles the Go binary for `linux/amd64` with symbols stripped, embedding the current version → `bin/server`
2. Builds the Vue frontend with `LINKSHORTENER_ENV=prod`, the provided `APP_BASE_URL`, and the current version baked in → `frontend/dist/`
3. Rsyncs `bin/server` to `/opt/linkshortener/server` on the target server
4. Rsyncs `frontend/dist/` to `/opt/linkshortener/frontend/` on the target server (old files are deleted)
5. Runs `sudo systemctl restart linkshortener` on the target server

> **Note:** Deployment causes a brief service interruption (~1–2 seconds) while the process restarts. Zero-downtime strategies (blue/green, rolling) are not implemented yet.
