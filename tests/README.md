# End-to-end tests

Playwright-based end-to-end browser tests for the Groceries web app.

These tests exercise the app the same way a user would: through a real
browser, against a live instance of the server backed by real Postgres and
Redis instances (the same ones defined in the root [docker-compose.yml](../docker-compose.yml)).

## Running

From the repository root:

```bash
mise run test:e2e
```

This will:

1. Start Postgres and Redis via `docker compose up -d`.
2. Start the compiled `./groceries` server binary (which applies database
   migrations automatically on startup).
3. Populate the database with deterministic seed data (`mise run seed`).
4. Install Playwright's Chromium browser and its dependencies.
5. Run the Playwright test suite in `tests/e2e/`.
6. Tear down the server process and docker compose services.

The suite currently runs Chromium only. Test files run serially (not in
parallel) because they mutate shared application state (the grocery list) in
the same way multiple users of the app would.

## Running Playwright directly

If the app and its dependencies are already running locally (e.g. via
`mise run default`) and seeded (`mise run seed`), you can iterate on the
tests without going through the full orchestration script:

```bash
cd tests
npm install
npx playwright test          # headless run
npx playwright test --ui     # interactive UI mode
```

Set `BASE_URL` if the app isn't running on the default
`http://localhost:3000`.
