# Financial System

REST API for membership fee collection: cities, members, recurring monthly charges, payments, and revenue dashboards.

Built as a Go monolith with Gin and PostgreSQL. There is no frontend in this repository.

## Features

- Email/password registration and login with bcrypt
- Stateless JWT authentication (`Authorization: Bearer <token>`)
- Cities and members (soft delete; unique phone per member)
- Fee types and recurring monthly charges (amounts stored in **cents**)
- Payments against a charge period, with overpayment protection
- Dashboard totals and revenue breakdowns by month, city, and fee type

## Stack

| Layer        | Choice                                      |
| ------------ | ------------------------------------------- |
| Language     | Go 1.26                                     |
| HTTP         | [Gin](https://github.com/gin-gonic/gin)     |
| Database     | PostgreSQL 16 (`pgx` driver)                |
| Auth tokens  | JWT HS256 (`golang-jwt/jwt/v5`)             |
| Passwords    | bcrypt                                      |
| Config       | Environment variables (optional `.env`)     |
| Deploy       | Docker multi-stage build + Compose          |

## How it works

Each **member** belongs to a **city**. A **fee type** (for example membership or parking) can be assigned as a **monthly charge** with an amount in cents and an optional end date. Staff **record payments** for a calendar month (`YYYY-MM`). The API refuses payments outside the charge window and refuses amounts that exceed the remaining balance for that period.

Money is always **integer cents** (e.g. `150000` = 1,500.00 in the local currency). Payment `period` is stored as the first day of the month.

```
City ──< Member ──< MonthlyCharge >── FeeType
                      │
                      └──< Payment (recorded_by → User)
```

Cities, members, fee types, and charges use **soft delete**. Payments are append-only.

## Project layout

```
.
├── main.go                 # Entry point
├── cmd/server.go           # DB connect, Gin, listen
├── config/                 # Env-based config
├── db/
│   ├── postgres.go
│   └── migrations/         # SQL schema (apply separately)
├── middleware/auth.go      # JWT Bearer middleware
├── utils/token/            # JWT generate / validate
└── modules/
    ├── router.go           # Route registration
    ├── user/               # Auth + user CRUD
    ├── city/
    ├── member/
    ├── feeTypes/
    ├── charge/             # Monthly charges
    ├── payment/
    └── dashboard/
```

Each module follows **handler → service → repository**.

## Prerequisites

- Go 1.26+
- PostgreSQL 16 (or Docker)
- Optional: [golang-migrate](https://github.com/golang-migrate/migrate) to apply SQL files

## Configuration

Copy the variables below into a `.env` file at the repo root, or export them in the shell. `godotenv` loads `.env` when present; otherwise system env is used.

| Variable             | Default          | Notes                                              |
| -------------------- | ---------------- | -------------------------------------------------- |
| `APP_NAME`           | `financial-system` | Logged at startup                                |
| `APP_ENV`            | `development`    | Set to `production` for Gin release mode           |
| `APP_PORT`           | `8080`           | HTTP listen port                                   |
| `DB_HOST`            | `localhost`      |                                                    |
| `DB_PORT`            | `5432`           |                                                    |
| `DB_USER`            | `postgres`       |                                                    |
| `DB_PASSWORD`        | `password`       |                                                    |
| `DB_NAME`            | `financial-v1`   |                                                    |
| `JWT_SECRET`         | _(empty)_        | **Required in real use.** HMAC signing key         |
| `JWT_EXPIRY_MINUTES` | `15`             | Integer minutes (not a duration string like `168h`) |

## Database

The API does **not** run migrations on startup. Apply the files in `db/migrations/` before the first request.

With [golang-migrate](https://github.com/golang-migrate/migrate):

```bash
migrate -path db/migrations -database "postgres://postgres:password@localhost:5432/financial-v1?sslmode=disable" up
```

Or with `psql`, in order:

```bash
psql "postgres://postgres:password@localhost:5432/financial-v1?sslmode=disable" -f db/migrations/000001_init.up.sql
psql "postgres://postgres:password@localhost:5432/financial-v1?sslmode=disable" -f db/migrations/000002_create_users_table.up.sql
psql "postgres://postgres:password@localhost:5432/financial-v1?sslmode=disable" -f db/migrations/000003_create_cities_table.up.sql
psql "postgres://postgres:password@localhost:5432/financial-v1?sslmode=disable" -f db/migrations/000004_create_members_table.up.sql
psql "postgres://postgres:password@localhost:5432/financial-v1?sslmode=disable" -f db/migrations/000005_create_fee_types_table.up.sql
psql "postgres://postgres:password@localhost:5432/financial-v1?sslmode=disable" -f db/migrations/000006_create_monthly_charges_table.up.sql
psql "postgres://postgres:password@localhost:5432/financial-v1?sslmode=disable" -f db/migrations/000007_create_payments_table.up.sql
```

## Run locally

```bash
createdb financial-v1   # if the database does not exist yet
# apply migrations (see above)

export JWT_SECRET=change-me
go run .
```

Health check:

```bash
curl http://localhost:8080/health
```

Expected: `{"status":"ok","database":"up"}`.

## Run with Docker

```bash
docker compose up --build
```

This starts PostgreSQL 16 and the API on port `8080`. Apply migrations against the Compose database (host port is not published by default — exec into the container or add a `ports` mapping for Postgres if you prefer running migrate from the host):

```bash
docker compose exec -T postgres psql -U postgres -d financial-v1 < db/migrations/000001_init.up.sql
# …repeat for 000002 through 000007
```

Set a strong `JWT_SECRET` in `docker-compose.yml` before any non-local use. The compose file’s `JWT_EXPIRY_MINUTES` must be an integer (the app parses it with `Atoi`; values like `168h` are ignored and the 15-minute default is used).

## Authentication

Public routes:

| Method | Path                     | Purpose              |
| ------ | ------------------------ | -------------------- |
| `GET`  | `/health`                | Liveness + DB ping   |
| `POST` | `/api/v1/auth/register`  | Create an account    |
| `POST` | `/api/v1/auth/login`     | Issue an access token |

Everything else under `/api/v1` requires:

```
Authorization: Bearer <access_token>
```

**Register** hashes the password with bcrypt and returns the user. It does **not** return a token — call login next.

**Login** verifies email and password, then returns a JWT whose only custom claim is `user_id`. Tokens are not stored server-side. There is no refresh or logout endpoint; clients discard the token, and expiry invalidates it.

Any authenticated user can call every protected route. There are no roles. The only extra rule is that a user cannot delete their own account.

```bash
# Register
curl -s http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Ada","email":"ada@example.com","password":"secret123"}'

# Login
curl -s http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ada@example.com","password":"secret123"}'
```

Login response:

```json
{
  "access_token": "<jwt>",
  "user": {
    "id": 1,
    "name": "Ada",
    "email": "ada@example.com",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

Password hashes are never included in JSON.

## API

Base URL: `http://localhost:8080/api/v1`

List endpoints share pagination: `page` (default `1`), `limit` (default `10`, max `100`). Response shape:

```json
{
  "data": [],
  "meta": { "page": 1, "limit": 10, "total": 0, "total_pages": 0 }
}
```

Errors are `{ "error": "..." }` with the matching HTTP status (`400`, `401`, `403`, `404`, `409`, `422`, `500`).

### Auth & users

| Method   | Path            | Auth | Body / query |
| -------- | --------------- | ---- | ------------ |
| `POST`   | `/auth/register` | No  | `name`, `email`, `password` (min 8) |
| `POST`   | `/auth/login`    | No  | `email`, `password` |
| `GET`    | `/me`            | Yes | Current user from the token |
| `GET`    | `/users`         | Yes | `page`, `limit`, `search` |
| `GET`    | `/users/:id`     | Yes | |
| `PUT`    | `/users/:id`     | Yes | `name`, `email` |
| `DELETE` | `/users/:id`     | Yes | `403` if you delete yourself |

### Cities

| Method   | Path           | Body / query |
| -------- | -------------- | ------------ |
| `POST`   | `/cities`      | `name`, `region` |
| `GET`    | `/cities`      | `page`, `limit`, `search` |
| `GET`    | `/cities/:id`  | |
| `PUT`    | `/cities/:id`  | `name`, `region` |
| `DELETE` | `/cities/:id`  | Soft delete. `409` if the city still has active members |

### Members

| Method   | Path            | Body / query |
| -------- | --------------- | ------------ |
| `POST`   | `/members`      | `full_name`, `phone`, `city_id` |
| `GET`    | `/members`      | `page`, `limit`, `search`, `city_id` |
| `GET`    | `/members/:id`  | |
| `PUT`    | `/members/:id`  | `full_name`, `phone`, `city_id` |
| `DELETE` | `/members/:id`  | Soft delete |

Phone must be unique among non-deleted members. `city_id` must exist (`422` otherwise).

### Fee types

| Method   | Path              | Body / query |
| -------- | ----------------- | ------------ |
| `POST`   | `/feeTypes`       | `name`, `description` (optional), `is_active` (optional, default `true`) |
| `GET`    | `/feeTypes`       | `page`, `limit`, `search` |
| `GET`    | `/feeTypes/:id`   | |
| `PUT`    | `/feeTypes/:id`   | `name`, `description`, `is_active` |
| `DELETE` | `/feeTypes/:id`   | Soft delete |

Inactive fee types cannot be used on new monthly charges.

### Monthly charges

| Method   | Path                   | Body / query |
| -------- | ---------------------- | ------------ |
| `POST`   | `/monthly-charges`     | `member_id`, `fee_type_id`, `amount_cents`, `start_date` (`YYYY-MM-DD`), optional `end_date` |
| `GET`    | `/monthly-charges`     | `page`, `limit`, `member_id`, `fee_type_id` |
| `GET`    | `/monthly-charges/:id` | |
| `PUT`    | `/monthly-charges/:id` | `amount_cents`, optional `end_date` |
| `DELETE` | `/monthly-charges/:id` | Soft delete |

A member may have only one **active** charge per fee type.

### Payments

| Method | Path               | Body / query |
| ------ | ------------------ | ------------ |
| `POST` | `/payments`        | `charge_id`, `period` (`YYYY-MM`), `amount_cents`, optional `note` |
| `GET`  | `/payments`        | `page`, `limit`, `member_id`, `charge_id` |
| `GET`  | `/payments/:id`    | |
| `GET`  | `/payments/summary`| **Required:** `charge_id`, `period` (`YYYY-MM`) |

`recorded_by` is taken from the JWT, not the request body.

Summary `status` is derived from paid vs charge amount for that month (unpaid / partial / paid).

### Dashboard

| Method | Path                            | Query |
| ------ | ------------------------------- | ----- |
| `GET`  | `/dashboard/summary`            | Totals: members, cities, collected, expected, outstanding |
| `GET`  | `/dashboard/monthly-revenue`    | `months` (default `12`, max `60`) |
| `GET`  | `/dashboard/revenue-by-city`    | |
| `GET`  | `/dashboard/revenue-by-fee-type`| |

## Example flow

```bash
TOKEN="<access_token from login>"
HDR=(-H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")

curl -s "${HDR[@]}" http://localhost:8080/api/v1/cities \
  -d '{"name":"Hargeisa","region":"Maroodi Jeex"}'

curl -s "${HDR[@]}" http://localhost:8080/api/v1/members \
  -d '{"full_name":"Amina Ali","phone":"634000000","city_id":1}'

curl -s "${HDR[@]}" http://localhost:8080/api/v1/feeTypes \
  -d '{"name":"Monthly membership","description":"Recurring fee"}'

curl -s "${HDR[@]}" http://localhost:8080/api/v1/monthly-charges \
  -d '{"member_id":1,"fee_type_id":1,"amount_cents":150000,"start_date":"2026-01-01"}'

curl -s "${HDR[@]}" http://localhost:8080/api/v1/payments \
  -d '{"charge_id":1,"period":"2026-03","amount_cents":150000,"note":"March"}'

curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/payments/summary?charge_id=1&period=2026-03"

curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/dashboard/summary
```

## License

Private / unlicensed unless you add one.
