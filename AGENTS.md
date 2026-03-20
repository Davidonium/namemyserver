# AGENTS.md - NameMyServer Development Guide

This document provides essential context for agents working on the namemyserver project. Use this as your reference for architecture, conventions, and development workflows.

## 1. Project Overview

**namemyserver** is a Go-based web application that generates memorable, human-friendly server names through a web UI and REST API.

### Purpose
Provides easily memorizable server names (e.g., "brave-mountain") for manual server naming tasks. Not intended for automated system provisioning.

### Tech Stack
- **Backend**: Go 1.26
- **Database**: SQLite with migrations (dbmate), separate read/write connection pools
- **Templates**: templ (type-safe HTML templates)
- **Frontend**: Vite, HTMX, Tailwind CSS v4, UmbrellaJS
- **Key Libraries**: sqlx, oapi-codegen, robfig/cron

## 2. Architecture & Core Concepts

### Domain Model

- **Pair**: An adjective-noun combination (e.g., "brave-mountain")
- **Bucket**: A named collection of pre-generated server names with specific filters
- **Generator**: Service that creates random pairs based on length constraints
- **Lifecycle**: Buckets transition through Active → Archived → Deleted (3 days after archive)

### Directory Structure

```
cmd/                         - Entry point (server, seed commands)
internal/
  ├── namemyserver/          - Core domain logic, types, business rules
  ├── server/                - HTTP handlers, routes, middleware, API layer
  │   └── api/               - Generated API code (oapi-codegen)
  ├── store/sqlitestore/     - Data access layer (SQLite implementation)
  ├── templates/             - Server-side rendering (templ)
  ├── bg/                    - Background task runner (cron-based)
  ├── vite/                  - Asset manifest handling
  ├── env/                   - Configuration structs
  └── dbtesting/             - Database test utilities
db/
  ├── migrations/            - Database schema migrations
  ├── seed/                  - Seed data (adjectives.txt, nouns.txt)
  └── schema.sql             - Complete database schema
frontend/                    - Vite-based assets, HTMX, Tailwind
  ├── src/
  │   ├── js/entries/        - JS entry points
  │   └── css/               - CSS (Tailwind)
  └── public/                - Static assets (favicon)
```

### Data Flow

```
API Request → Route → Handler → Business Logic → Store → Database
                        ↓
                     Response

Web Request → Route → Handler → Generator/Store → templ Template → HTTP Response
```

## 3. API Design Conventions

### Critical Design Principles

1. **Plural nouns for resources**: Use `/buckets`, `/pairs`, not `/bucket`, `/pair`
2. **snake_case for JSON fields**: `created_at`, `length_mode`, `remaining_pairs`, `archived_at`
3. **Versioning**: All APIs under `/api/v1alpha1/`
4. **OpenAPI-first**: Spec in `openapi.yaml` is the source of truth, code is generated
5. **Error handling**: RFC 7807 Problem Details for client errors, generic Error schema for 5xx



## 4. Name Generation Rules

### Format
`{adjective}-{noun}` (e.g., "brave-mountain", "agile-abyss")

### Validation (RFC 1123 DNS Subdomain Compliant)
- Lowercase alphanumeric + hyphens only
- Must start and end with alphanumeric character
- Maximum 63 characters total
- Regex: `^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`

### Seed Data
- **Adjectives**: 347 words (able, active, adept, admiring, etc.)
- **Nouns**: 583 words (abyss, acorn, adventure, aether, etc.)
- **Combinations**: ~202,301 unique pairs

### Generation Filters
- `length_mode`: "exactly" or "upto"
- `length`: Character constraint for total name length
- If `length_enabled` is false, any valid pair is returned

## 5. Database Schema & Connection Pools

### Connection Pool Architecture

The database layer uses **separate read and write connection pools** to optimize SQLite performance and prevent lock contention:

- **Write Pool** (`DBPool.Write()`): Single connection for all INSERT, UPDATE, DELETE, and transaction operations
  - `MaxOpenConns = 1` (SQLite writes are serialized)
  - Uses `_txlock=immediate` for immediate transaction locking
- **Read Pool** (`DBPool.Read()`): Multiple connections for SELECT operations (default: max 4 or NumCPU, whichever is greater)
  - `MaxOpenConns = max(4, runtime.NumCPU())`
  - Standard transaction mode

### Routing Read vs Write Operations

All store methods must explicitly choose the correct pool:

```go
// Write operations
func (s *BucketStore) Create(ctx context.Context, b *Bucket) error {
    r, err := s.db.Write().NamedExecContext(ctx, createBucketSQL, args)
    // ...
}

// Read operations
func (s *BucketStore) OneByID(ctx context.Context, id int32) (Bucket, error) {
    stmt, err := s.db.Read().PrepareNamedContext(ctx, oneByIDSQL)
    // ...
}

// Transactions
func (s *BucketStore) PopName(ctx context.Context, b Bucket) (string, error) {
    err := s.db.Write().WithTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx *sqlx.Tx) error {
        // transaction logic
    })
    // ...
}
```

### SQLite Configuration (Both Pools)

Both read and write pools share common PRAGMA settings:

```
_journal_mode=WAL          - Write-Ahead Logging for better concurrency
_synchronous=NORMAL        - Balance between safety and performance
_busy_timeout=5000ms       - Wait up to 5 seconds for locks
_cache_size=256MB          - L1 page cache
_foreign_keys=ON           - Enforce foreign key constraints
_temp_store=MEMORY         - Use memory for temp tables
_mmap_size=256MB           - Memory-mapped I/O for faster reads
```

### Tables

```sql
-- Word pools for pair generation
nouns
  - id (PK)
  - value (TEXT, UNIQUE)
  - from_seed (INT) -- 1 if from seed data, 0 if custom added
  - created_at (TIMESTAMP)
  - updated_at (TIMESTAMP)

adjectives
  - id (PK)
  - value (TEXT, UNIQUE)
  - from_seed (INT)
  - created_at (TIMESTAMP)
  - updated_at (TIMESTAMP)

-- Bucket management
buckets
  - id (PK)
  - name (TEXT, UNIQUE)
  - description (TEXT)
  - cursor (INT) -- tracks position for next pop
  - created_at (TIMESTAMP)
  - updated_at (TIMESTAMP)
  - archived_at (TIMESTAMP, NULL) -- NULL = active, set = archived

-- Bucket contents (pre-generated names)
bucket_values
  - id (PK)
  - bucket_id (FK → buckets.id, CASCADE delete)
  - order_id (INT) -- determines pop order
  - value (TEXT) -- the actual name
  - created_at (TIMESTAMP)
  - updated_at (TIMESTAMP)
```

### Important Details
- **Bucket cursor**: Tracks the next position when popping names. Increments with each pop until all names exhausted.
- **Archived buckets**: Can only be read (GET operations). Pop is forbidden (409 Conflict).
- **Bucket deletion**: Background task runs hourly, removes buckets archived >3 days ago.
- **Pool management**: `DBPool.Close()` closes both read and write connections. Always defer cleanup.

## 6. Development Workflow

### Build System (`./please` script)

The `please` script is the primary build tool. Key commands:

```bash
./please build              # Production binary (build/namemyserver)
./please generate           # Run templ + oapi-codegen
./please test               # Run all tests
./please coverage           # Generate coverage report
./please lint               # golangci-lint style checks
./please format             # Auto-fix + templ fmt
./please dev                # Hot reload server (Air)
./please dev:frontend       # Vite watch mode
./please db:migrate         # Apply pending migrations
./please db:reset           # DROP database, recreate, migrate (dev only!)
./please db:seed            # Seed adjectives/nouns tables
./please docker             # Build Docker image
```

### Code Generation

#### templ - Type-Safe HTML Templates
```bash
go tool templ generate
```
- Converts `.templ` files → `*_templ.go` Go code
- Type-safe templates with compile-time checking
- Generate with: `./please generate` or via Air on `.templ` file changes

#### oapi-codegen - API Generation
```bash
go tool oapi-codegen -config oapicodegen.config.yaml openapi.yaml
```
- Input: `openapi.yaml` (source of truth)
- Output: `internal/server/api/namemyserver_api.gen.go`
- Generates: data models, HTTP server interface, embedded spec
- Config: `oapicodegen.config.yaml` enables std-http-server, models, strict-server

### Live Reload Setup

#### Backend
- **Tool**: Air (.air.toml)
- **Watches**: `.go`, `.templ`, `.html` files
- **Excludes**: `*_templ.go`, `*_test.go`
- **Run**: `./please dev`
- **Output**: `build/namemyserver-dev`

#### Frontend
- **Tool**: Vite
- **Watches**: Everything in `frontend/src/`
- **Build output**: `frontend/dist/`
- **Run**: `./please dev:frontend`
- **Dev mode origin**: http://127.0.0.1:8080

##### Frontend Deesign

<frontend_aesthetics>
You tend to converge toward generic, "on distribution" outputs. In frontend design, this creates what users call the "AI slop" aesthetic. Avoid this: make creative, distinctive frontends that surprise and delight. Focus on:

Typography: Choose fonts that are beautiful, unique, and interesting. Avoid generic fonts like Arial and Inter; opt instead for distinctive choices that elevate the frontend's aesthetics.

Color & Theme: Commit to a cohesive aesthetic. Use CSS variables for consistency. Dominant colors with sharp accents outperform timid, evenly-distributed palettes. Draw from IDE themes and cultural aesthetics for inspiration.

Motion: Use animations for effects and micro-interactions. Prioritize CSS-only solutions for HTML. Use Motion library for React when available. Focus on high-impact moments: one well-orchestrated page load with staggered reveals (animation-delay) creates more delight than scattered micro-interactions.

Backgrounds: Create atmosphere and depth rather than defaulting to solid colors. Layer CSS gradients, use geometric patterns, or add contextual effects that match the overall aesthetic.

Avoid generic AI-generated aesthetics:
- Overused font families (Inter, Roboto, Arial, system fonts)
- Clichéd color schemes (particularly purple gradients on white backgrounds)
- Predictable layouts and component patterns
- Cookie-cutter design that lacks context-specific character

Interpret creatively and make unexpected choices that feel genuinely designed for the context. Vary between light and dark themes, different fonts, different aesthetics. You still tend to converge on common choices (Space Grotesk, for example) across generations. Avoid this: it is critical that you think outside the box!
</frontend_aesthetics>

#### Running Both
```bash
# Terminal 1
./please dev:frontend

# Terminal 2
./please dev
```

## 7. Frontend Integration

### Technologies

- **HTMX v2**: Dynamic interactions, form submissions, server-driven UI
- **Tailwind CSS v4**: Utility-first styling via Vite plugin
- **UmbrellaJS** : Lightweight DOM manipulation
- **Vite v7**: Fast build tool with HMR
- **templ**: Server-side rendering of HTML components

### Asset Pipeline

#### Development Mode
- Vite dev server at `http://127.0.0.1:8080`
- Assets NOT embedded, served from filesystem
- Live module reloading
- Config: `ASSETS_MANIFEST_USE=false`

#### Production Mode
- Static files built into `frontend/dist/`
- Assets embedded in binary via `embed.go`
- Manifest file for asset versioning (`manifest.json`)
- Config: `ASSETS_MANIFEST_USE=true`, `ASSETS_MANIFEST_FS=embed`

### Asset Configuration Variables

```
ASSETS_ROOT_URL          - Base URL for asset requests (e.g., /static)
ASSETS_MANIFEST_USE      - Enable manifest mode (default: false)
ASSETS_MANIFEST_LOCATION - Path to manifest.json
ASSETS_MANIFEST_WATCH    - Auto-reload manifest on changes (dev only)
ASSETS_MANIFEST_FS       - "os" (filesystem) or "embed" (embedded)
```

### Pages & Routes

- **GET /**: Home page (name generator form)
- **GET /buckets**: Bucket list page
- **GET /buckets/create**: Create bucket form
- **GET /buckets/{id}**: Bucket details page
- **GET /stats**: Statistics dashboard
- **GET /config/stats**: Stats partial (config display)
- **POST /generate**: Generate name (HTMX)
- **POST /buckets**: Create bucket submit
- **POST /buckets/{id}/pop**: Pop name from bucket (HTMX)
- **POST /buckets/{id}/archive**: Archive bucket
- **POST /buckets/{id}/recover**: Un-archive bucket

## 8. Background Tasks

### Architecture

- **Framework**: robfig/cron/v3 (cron expression scheduling)
- **Runner**: Starts on server startup (`bg.NewRunner`)
- **Error handling**: Logged but non-fatal, task continues running

### Task Definition Pattern

```go
func removeArchivedBucketsTask(
    logger *slog.Logger,
    bucketStore namemyserver.BucketStore,
) func(context.Context) error {
    return func(ctx context.Context) error {
        // task logic
        return nil
    }
}
```

### Naming Convention

- **Task names must be snake_case** (enforced at runtime with regex check)
- Pattern: `^[a-z0-9]+(_[a-z0-9]+)*$`
- Panic if invalid name provided

### Adding a New Task

1. Create task function in `internal/bg/` returning `func(context.Context) error`
2. Register in `runner.go` `setup()` method:
   ```go
   r.cron.AddFunc("0 * * * *", r.task("task_name", taskFunc))
   ```
3. Ensure task name is snake_case
4. Task receives logger and dependencies via closure

## 9. Configuration

### Environment Variables

```
LISTEN_ADDR              - Server listen address (default: :8080)
DATABASE_URL             - SQLite connection string (required)
                           Format: sqlite:./var/namemyserver.db
DEBUG                    - Enable debug mode, expose error details (default: false)
LOG_FORMAT               - "text" or "json" (default: text)
LOG_LEVEL                - slog level: debug, info, warn, error (default: info)
ASSETS_ROOT_URL          - Base URL for assets (required)
                           Format: /static or https://cdn.example.com
ASSETS_MANIFEST_USE      - Enable manifest mode (default: false)
ASSETS_MANIFEST_LOCATION - Path to manifest.json file
ASSETS_MANIFEST_WATCH    - Watch manifest for changes (default: false)
ASSETS_MANIFEST_FS       - "os" (filesystem) or "embed" (embedded) (default: os)
```

### Environment Loading

1. Check for `.env` file in project root (via godotenv)
2. Parse variables into `internal/env/Config` struct
3. Panic if required variables missing

### Example `.env`

```bash
LISTEN_ADDR=:8080
DATABASE_URL=sqlite:./var/namemyserver.db
DEBUG=true
LOG_FORMAT=text
LOG_LEVEL=info
ASSETS_ROOT_URL=http://localhost:5173
ASSETS_MANIFEST_USE=false
ASSETS_MANIFEST_WATCH=false
ASSETS_MANIFEST_FS=os
```

## 10. Code Conventions

### Package Organization

| Package | Purpose |
|---------|---------|
| `internal/namemyserver` | Domain types, interfaces, business logic |
| `internal/server` | HTTP handlers, routes, middleware |
| `internal/server/api` | Generated API types (from openapi.yaml) |
| `internal/store/sqlitestore` | SQLite implementation of domain interfaces |
| `internal/templates` | templ server-side templates |
| `internal/bg` | Background task runner and tasks |
| `internal/vite` | Asset manifest loading utilities |
| `internal/env` | Configuration structs with env tags |

### Naming Patterns

#### Cron Tasks
- Must be **snake_case**: `remove_archived_buckets`, `cleanup_stale_buckets`
- Enforced via regex at runtime

#### HTTP Handlers
- **Pattern**: `{resource}{action}Handler`
- **Examples**: `homeHandler`, `bucketListHandler`, `bucketCreateSubmitHandler`
- **Signature**: `func(...) func(http.ResponseWriter, *http.Request) error`

#### Stores
- **Interface**: Defined in domain package (`internal/namemyserver`)
- **Examples**: `BucketStore`, `PairStore`
- **Implementation**: In `internal/store/sqlitestore/`

#### Templates
- **PascalCase**: `HomePage`, `BucketListPage`, `GeneratePartial`
- **Suffix**: `Page` for full pages, `Partial` for fragments

#### Validation Functions
- **Pattern**: `Validate{Something}`
- **Examples**: `ValidateName`, `ValidateNameSegment`
- **Returns**: `bool`

#### HTML Form Fields & Query String Variables
- **Convention**: **snake_case** for all form `name` attributes and query string parameters
- **Examples**: `length_enabled`, `filter_length_mode`, `filter_length_value`, `length_mode`
- **Backend Reading**: Use `r.FormValue("field_name")` with snake_case to match form attributes
- **Rationale**: Consistency with REST API conventions and improved readability
- **Template Example**:
  ```templ
  <input type="radio" name="length_mode" value="upto" />
  <input name="filter_length_value" type="range" />
  ```
- **Handler Example**:
  ```go
  lengthMode := r.FormValue("length_mode")
  filterValue := r.FormValue("filter_length_value")
  ```

### Error Handling

#### Web Handlers
```go
type appHandlerFunc func(http.ResponseWriter, *http.Request) error

func homeHandler(store Store) appHandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        // If you return an error, WebErrorHandler renders error page
        return nil
    }
}
```

#### API Handlers
- Use generated types from oapi-codegen
- Return RFC 7807 Problem Details for client errors (4xx)
- Return Error schema for server errors (5xx)
- HTTP middleware handles serialization

#### Middleware Pattern
```go
type MiddlewareFunc func(http.Handler) http.Handler

// Chain multiple middleware
c := chainMiddleware([]MiddlewareFunc{
    viteMiddleware(assets),
})

// Application middleware with error handling
app := appMiddleware(logger, errorHandler)

// Register route with middleware
m.Handle("GET /path", c(app(handler)))
```

### Logging

- **Logger**: slog (structured logging)
- **Pattern**: Pass logger to services/handlers
- **Scoping**: Add context with `.With()`: `logger.With("service", "generator")`
- **Levels**: debug, info, warn, error

Example:
```go
logger.Info("starting task",
    slog.String("name", "remove_archived_buckets"),
)
logger.Error("failure running task",
    slog.Any("err", err),
    slog.String("task", name),
)
```

## 11. Testing

### Test Structure
- **Location**: `*_test.go` files alongside source code
- **Framework**: stretchr/testify for assertions
- **Utilities**: `internal/dbtesting` for database setup/teardown

### Current Tests
- **validation_test.go**: Name validation rules (RFC 1123 compliance)
- **bucket_store_test.go**: Data access layer (Create, Read, Update, Delete, Pop)
- **assets_test.go**: Vite manifest loading

### Running Tests
```bash
./please test              # Run all tests
./please coverage          # Generate coverage.out
./please see_coverage      # Open coverage report in browser
```

### Coverage Output
- File: `build/coverage.out`
- Generated by: `./please coverage`

### Database Testing
- Use `internal/dbtesting/run.go` for test database setup
- Tests should clean up after themselves (transactions, drops)

## 12. Docker Deployment

### Multi-Stage Build

#### Stage 1: Frontend
```dockerfile
FROM node:24-slim AS frontend
# Build Vite assets → frontend/dist/
```

#### Stage 2: Builder
```dockerfile
FROM golang:1.25 AS builder
# Compile Go binary → build/namemyserver
# Embed frontend dist + migrations
```

#### Stage 3: Final
```dockerfile
FROM gcr.io/distroless/static-debian13:nonroot
# Copy binary only (distroless = minimal attack surface)
```

### Key Points
- Frontend assets embedded in binary via `embed.go`
- Database migrations included in binary
- Binary runs as non-root user (distroless requirement)
- Entrypoint: `/app/namemyserver server`
- No shell, no package manager (security)

### Building Image
```bash
./please docker
# Creates: davidonium/namemyserver:0.1.0
# Tags: davidonium/namemyserver:latest
```

## 13. Important Files Reference

| File | Purpose |
|------|---------|
| `openapi.yaml` | API specification (source of truth) |
| `oapicodegen.config.yaml` | API code generation configuration |
| `embed.go` | Embeds migrations + frontend dist into binary |
| `db/schema.sql` | Complete database schema (reference only) |
| `db/migrations/20240609195352_init.sql` | Initial database schema |
| `.air.toml` | Hot reload configuration (backend) |
| `frontend/vite.config.ts` | Frontend build configuration |
| `please` | Build task runner (bash script) |
| `.golangci.yml` | Linter configuration |

## 14. Common Development Tasks

### Understanding Handler Types

There are two distinct handler patterns in this project:

#### Web Handlers (Server-Side Rendering)
- **Pattern**: Functions returning `appHandlerFunc`
- **Location**: `internal/server/*_handlers.go`
- **Registration**: Manual route registration in `routes.go`
- **Response**: HTML rendered via templ templates
- **Example**: `homeHandler`, `bucketListHandler`, `bucketCreateSubmitHandler`

#### API Handlers (JSON/REST)
- **Pattern**: Methods on `Handlers` struct implementing `StrictServerInterface`
- **Location**: `internal/server/api/api_handlers.go`
- **Registration**: Automatic via generated `StrictServerInterface`
- **Response**: JSON responses with proper status codes
- **Example**: `GenerateName`, `CreateBucket`, `ListBuckets`, `PopBucketName`

### Adding a New Web Endpoint (Server-Side Rendering)

1. **Create handler function** in `internal/server/`
   ```go
   func myPageHandler(store Store) appHandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) error {
           // fetch data, call business logic
           return component(w, r, http.StatusOK, templates.MyPage(vm))
       }
   }
   ```

2. **Create templ template** in `internal/templates/`
   ```templ
   package templates
   templ MyPage(vm MyPageViewModel) {
       <h1>{ vm.Title }</h1>
   }
   ```

3. **Generate templ code**
   ```bash
   ./please generate  # or let Air auto-generate on file save
   ```

4. **Register route** in `internal/server/routes.go`
   ```go
   m.Handle("GET /my-page", c(app(myPageHandler(svcs.Store))))
   ```

5. **Middleware**: Web handlers use chain + app middleware for error handling
   ```go
   c := chainMiddleware([]MiddlewareFunc{
       viteMiddleware(svcs.Assets),
   })
   app := appMiddleware(svcs.Logger, WebErrorHandler(svcs.Logger, svcs.Config.Debug))
   m.Handle("GET /path", c(app(handler)))
   ```

### Adding a New API Endpoint

1. **Update OpenAPI spec** (`openapi.yaml`)
   ```yaml
   paths:
     /v1alpha1/new-resource:
       post:
         operationId: createNewResource
         requestBody: {...}
         responses: {...}
   ```

2. **Generate API code**
   ```bash
   ./please generate
   ```
   This generates the request/response types and `StrictServerInterface` that your handlers must implement.

3. **Implement handler method on `Handlers` struct** in `internal/server/api/api_handlers.go`
   ```go
   func (s *Handlers) CreateNewResource(
       ctx context.Context,
       request CreateNewResourceRequestObject,
   ) (CreateNewResourceResponseObject, error) {
       // Implement business logic
       // Return appropriate response type (e.g., CreateNewResource201Response{})
       // Return error if operation fails
   }
   ```
   - Method name matches the `operationId` from OpenAPI spec
   - Request object auto-generated from spec
   - Return type must be one of the generated `ResponseObject` types
   - Error handling: return error for 5xx, return proper response object for 4xx

4. **Register route** in `internal/server/routes.go`
   - API routes are automatically registered via the `StrictServerInterface` implementation
   - No manual route registration needed (handled by generated code)

5. **Test** and verify API contract

### Adding a New templ Template

1. **Create template file** `internal/templates/my_page.templ`
   ```templ
   package templates

   templ MyPage(vm MyPageViewModel) {
       <h1>{ vm.Title }</h1>
   }
   ```

2. **Generate code**
   ```bash
   ./please generate  # or let Air auto-generate
   ```

3. **Use in handler** `internal/server/handlers.go`
   ```go
   func myHandler(...) appHandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) error {
           return component(w, r, http.StatusOK, templates.MyPage(vm))
       }
   }
   ```

4. **Format templates**
   ```bash
   ./please format  # includes templ fmt
   ```

### Adding a Background Task

1. **Create task file** `internal/bg/my_task.go`
   ```go
   func myTask(
       logger *slog.Logger,
       store SomeStore,
   ) func(context.Context) error {
       return func(ctx context.Context) error {
           logger.Info("running my task")
           // task logic
           return nil
       }
   }
   ```

2. **Register in runner** `internal/bg/runner.go`
   ```go
   func (r *Runner) setup() {
       r.cron.AddFunc(
           "0 * * * *",  // hourly
           r.task("my_task", myTask(r.logger, r.store)),
       )
   }
   ```

3. **Ensure snake_case** task name (enforced!)

4. **Test** with context and logger

### Modifying Seed Data

1. **Edit seed files**
   ```bash
   vim db/seed/adjectives.txt
   vim db/seed/nouns.txt
   ```

2. **Run seed command**
   ```bash
   ./please db:seed
   ```

3. **Seed behavior**
   - **Adds**: New words not in database
   - **Removes**: Words removed from seed files (if `from_seed=1`)
   - **Preserves**: Custom entries added manually (`from_seed=0`)

### Running Tests Locally

```bash
# All tests
./please test

# With coverage
./please coverage

# View coverage in browser
./please see_coverage
```

### Linting and Formatting

```bash
# Check code quality
./please lint

# Auto-fix + format
./please format

# Format just templates
go tool templ fmt ./internal/templates
```

## 15. Useful Debugging Tips

### Enable Debug Mode
```bash
DEBUG=true ./please dev
# Errors include internal type/message in responses
```

### Check Database State
```bash
sqlite3 var/namemyserver.db
sqlite> SELECT * FROM buckets;
sqlite> SELECT COUNT(*) FROM bucket_values WHERE bucket_id=1;
```

### View Server Logs
- Set `LOG_LEVEL=debug` for verbose output
- Set `LOG_FORMAT=json` for structured logging to parse

### Check Generated Code
- `internal/server/api/namemyserver_api.gen.go` - API types
- `internal/templates/*_templ.go` - Template code
- These are safe to inspect but DON'T EDIT (regenerate instead)

### Asset Debugging
- Check `ASSETS_MANIFEST_USE` setting
- Verify `frontend/dist/manifest.json` exists
- Check browser console for 404s on assets
- Ensure `ASSETS_ROOT_URL` matches request prefix

---

## Quick Reference

### Start Development
```bash
./please install_tools      # One-time setup
./please db:reset           # Fresh database
./please db:seed            # Load seed data
./please dev:frontend &     # Start Vite (background)
./please dev                # Start server (foreground, hot reload)
```

### Common Commands
```bash
./please build              # Production build
./please test               # Run tests
./please lint && ./please format  # QA checks
./please docker             # Build container
```

### Database
```bash
./please db:migrate         # Apply pending migrations
./please db:reset           # Wipe and recreate (DEV ONLY)
./please db:seed            # Load/update seed words
```

---

**Last Updated**: January 2026
**Go Version**: 1.25
**Status**: Active Development (v0.1.0-alpha)
