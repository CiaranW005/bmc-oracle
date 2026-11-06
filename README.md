# BMC Oracle Server

The Oracle server connects the [BMC](../README.md) browser dashboard to the ESP32. It provides authenticated HTTP endpoints, maintains the device WebSocket, forwards commands, and publishes live status.

## Implementation

`cmd/bmc-oracle/main.go` loads configuration, starts acknowledgement/response tracking and telemetry polling, registers HTTP routes, and listens on port 8080.

The device manager holds one current ESP32 connection. A newly authenticated connection replaces the previous one, and cleanup only removes a connection if it is still current. Separate locks protect connection selection, packet writes, and shared status.

Commands are registered with both delivery and response trackers before sending. An acknowledgement confirms receipt; a response is matched using `reply_to`. Duplicate device messages are acknowledged again without resolving them twice.

Wake has a visible lifecycle: `sending`, `booting` after acknowledgement, then `online` or `failed`. The API waits up to 55 seconds for the result, leaving room beyond the firmware's 50-second deadline. Restart and shutdown wait five seconds; a response timeout returns `no_failure_reported`, which is not confirmation that the operation completed.

Telemetry polling is driven by a 1.5-second ticker, with polls performed sequentially and a three-second response timeout. Valid telemetry marks the BMC available and uses the reported `pc` value for PC availability. Failed or invalid polls mark both unavailable. `lastSeen` retains the most recent successful BMC observation.

Status subscribers get an initial snapshot and subsequent updates. Each has a one-entry channel; an older queued snapshot is replaced rather than blocking publication. The HTTP handler sends these snapshots as server-sent events and sends keepalive comments every 15 seconds.

## Authentication

Browser access supports a bcrypt password or an allowed Google account. Both create the same eight-hour, non-sliding session. Session tokens live in HTTP-only cookies; only their SHA-256 hashes are stored in memory. Restarting the server invalidates sessions.

Production uses a Secure `__Host-` cookie. Development permits HTTP cookies. Google credentials are validated against the configured client ID and checked against an explicit subject-ID allowlist.

The device endpoint uses a separate bearer token. Production ESP32 firmware also presents a client certificate to the TLS endpoint; the Go HTTP listener itself does not configure TLS or client-certificate verification.

## Configuration

The server optionally loads `.env` from its working directory. Existing process environment values take precedence.

| Variable | Purpose |
| --- | --- |
| `BMC_ENV` | Required: `development` or `production` |
| `ESP32_DEVICE_TOKEN` | Required bearer token shared with the ESP32 |
| `BMC_PASSWORD_HASH` | Required bcrypt hash for password login |
| `GOOGLE_CLIENT_ID` | Required Google client ID |
| `GOOGLE_ALLOWED_SUBS` | Comma-separated allowed Google subject IDs; empty denies all Google accounts |

Generate a password hash from this directory:

```bash
go run ./cmd/password
```

Store the printed hash in `BMC_PASSWORD_HASH`. The parent repo includes `deployment/env/production.env.example`; Google settings must be supplied in addition to its values.

## Run and build

Use Go 1.26.0 and the parent BMC workspace, whose `go.work` supplies the local protocol module:

```bash
# From oracle-server, after configuring the environment:
go run ./cmd/bmc-oracle

go build -o bmc-oracle ./cmd/bmc-oracle
```

## HTTP routes

| Route | Authentication | Purpose |
| --- | --- | --- |
| `/device` | Device bearer token | ESP32 WebSocket connection |
| `POST /api/login` | Password in JSON body | Create browser session |
| `POST /api/auth/google` | Google credential in JSON body | Create browser session |
| `POST /api/logout` | Session cookie | End session |
| `GET /api/session` | Session cookie | Return 204 for a valid session |
| `GET /api/status/events` | Session cookie | Stream status snapshots |
| `POST /api/wake` | Session cookie | Wake PC and await startup result |
| `POST /api/reset` | Session cookie | Request graceful reboot |
| `POST /api/power_off` | Session cookie | Request graceful shutdown |

## Deployment

The parent repository builds this server and the frontend together. Its deployment script installs the executable at `/usr/local/bin/bmc-oracle` and a systemd unit running as `ubuntu`, with environment values loaded from `/home/ubuntu/.config/bmc-oracle/.env`. Caddy provides the external routing/TLS layer; its referenced configuration is absent from this checkout.
