# Pratello

Banco de dados, caso de teste, e api esta definido em docs!

A full-stack web application with a **Go (standard `net/http`) backend**, **Svelte frontend**, and **Tailwind CSS**, managed with **Nix flakes**.

---

## 📁 Project Structure

```text
pratello/
├── flake.nix                # Nix environment (Go, Node.js, Docker, Java 21, Gradle)
├── go.mod                   # Go 1.22 module definition
├── package.json             # Root workspace scripts (npm run dev, test:e2e, server)
├── README.md                # Project documentation
├── server/
│   └── main.go              # Go standard net/http server & API routes
├── frontend/                # Svelte + Tailwind + Vite
│   ├── index.html           # HTML template
│   ├── package.json         # Frontend dependencies
│   ├── vite.config.js       # Vite configuration with API proxy
│   ├── tailwind.config.js   # Tailwind configuration
│   ├── postcss.config.js    # PostCSS configuration
│   └── src/
│       ├── app.css          # Tailwind CSS directives
│       ├── main.js          # Svelte mount entrypoint
│       ├── App.svelte       # Root component
│       └── Home.svelte      # Main / restaurant page component
└── e2e/                     # Java E2E test suite (Gradle + Playwright)
    ├── settings.gradle.kts  # Gradle settings
    ├── build.gradle.kts     # Playwright & JUnit 5 dependencies
    └── src/test/java/org/pratello/
        └── PratelloE2ETest.java
```

---

## 🚀 Getting Started with Nix

### 1. Enter the Nix Development Shell
```bash
nix develop
```
This loads **Go**, **Node.js (v22)**, and **Docker** into your shell.

---

### 2. Development Mode (Recommended)

Run the backend and frontend concurrently for hot-reloading:

**Terminal 1 (Backend):**
```bash
go run ./server
```
*(Runs at `http://localhost:8080`)*

**Terminal 2 (Frontend with Vite + HMR):**
```bash
cd frontend
npm install
npm run dev
```
*(Runs at `http://localhost:5173` with live-reload and automatic proxying of `/api` requests to Go)*

---

### 3. Run the João Burger development demo

Point `DATABASE_URL` at a dedicated development PostgreSQL database. Every `npm run server` boot drops and recreates the active schema, resets the served `media/` directory, and loads the João Burger fixture. The database role needs permission to drop and create that schema. All demo accounts use the password `password`.

```bash
DATABASE_URL='postgres://localhost/pratello_dev?sslmode=disable' npm run server
```

`npm run server` reuses a running local Redis or starts the `pratello-dev-redis` container with Docker. It waits for Redis at `redis://localhost:6379/0` and starts the API with `AUTH_FIXED_CODE=1`. Request a code through the app, then enter `123456`; no email or WhatsApp message is sent. To stop the Redis container started by this command, run `docker stop pratello-dev-redis`.

The API serves media from `./media` by default. Set `MEDIA_DIR` to a different directory named `media` to use that location for the demo. For a regular startup, run `go run ./server/cmd/api`.

---

## 🌐 Routes Overview

| Route | Description |
| :--- | :--- |
| `/` | Pratello Homepage (rendered by [Home.svelte](file:///programs/codes/projects/pratello/frontend/src/Home.svelte)) |
| `/{restaurant}` | Dynamic restaurant view (e.g. `/bobpizzeria`, `/deliciadeprazeres`) |
| `/api/health` | Healthcheck endpoint (`{"status":"ok"}`) |
| `/api/restaurants/{restaurant}` | JSON restaurant details served by Go `net/http` |

---

## Backend tests (Go)

From the repository root, start PostgreSQL and Redis instances for testing and set their connection URLs:

```bash
export AUTH_TEST_DATABASE_URL='postgres://localhost/pratello_test?sslmode=disable'
export AUTH_TEST_REDIS_URL='redis://127.0.0.1:6379/0'
nix develop -c node scripts/test-backend-cases.mjs
```

The report groups case IDs by backend module. Planned and skipped cases appear separately from executed cases. The PostgreSQL role needs permission to create test schemas. Without the test URLs, some integration tests are skipped.

`XFAIL` means a test reproduced a known failure. `XPASS` means the expected behavior now works; it fails the run so the test can be converted to a normal regression test.

To see each Go test function and subtest in the terminal, use the same environment variables and run:

```bash
nix develop -c go test -v -count=1 ./server/...
```

## OpenAPI reference

Open [`docs/openapi.yaml`](docs/openapi.yaml) in your editor to read the source. To browse the same specification in Swagger UI, run this from the repository root and open <http://localhost:8081>:

```bash
docker run --rm -p 127.0.0.1:8081:8080 \
  -v "$PWD/docs:/spec:ro" \
  -e SWAGGER_JSON=/spec/openapi.yaml \
  swaggerapi/swagger-ui
```

## 🧪 Running E2E Tests (Java + Playwright + Gradle)

With the app running (e.g. `npm run dev` and `npm run server`):

```bash
# Run tests with Gradle
gradle -p e2e test

# Or via npm script from root
npm run test:e2e

# Run with browser window visible (headed mode)
HEADLESS=false gradle -p e2e test
```
