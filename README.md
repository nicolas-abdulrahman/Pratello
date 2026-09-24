# Pratello

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

### 3. Production Build

To build the static Svelte assets and serve them directly from the Go server:

```bash
# 1. Build frontend
npm install
npm run dev 
cd ..

# 2. Start Go server
npm run server
```

---

## 🌐 Routes Overview

| Route | Description |
| :--- | :--- |
| `/` | Pratello Homepage (rendered by [Home.svelte](file:///programs/codes/projects/pratello/frontend/src/Home.svelte)) |
| `/{restaurant}` | Dynamic restaurant view (e.g. `/bobpizzeria`, `/deliciadeprazeres`) |
| `/api/health` | Healthcheck endpoint (`{"status":"ok"}`) |
| `/api/restaurants/{restaurant}` | JSON restaurant details served by Go `net/http` |

---

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
