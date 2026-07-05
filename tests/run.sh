#!/usr/bin/env bash
# Spins up the app's dependencies and a live instance of the server, runs the
# Playwright end-to-end test suite against it, and tears everything down
# again. Intended to be invoked via `mise run test:e2e`.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_PID=""
COMPOSE_STARTED=0

cleanup() {
  if [ -n "$APP_PID" ]; then
    echo "==> Stopping application server"
    kill "$APP_PID" 2>/dev/null || true
    wait "$APP_PID" 2>/dev/null || true
  fi

  if [ "$COMPOSE_STARTED" = "1" ]; then
    echo "==> Stopping docker compose services"
    docker compose down
  fi
}
trap cleanup EXIT

wait_for_port() {
  local host="$1" port="$2" name="$3"
  for _ in $(seq 1 60); do
    if (exec 3<>"/dev/tcp/${host}/${port}") 2>/dev/null; then
      exec 3>&- 3<&-
      return 0
    fi
    sleep 1
  done

  echo "Timed out waiting for ${name} on ${host}:${port}" >&2
  return 1
}

echo "==> Starting dependencies (postgres, redis)"
docker compose up -d
COMPOSE_STARTED=1

wait_for_port 127.0.0.1 5432 postgres
wait_for_port 127.0.0.1 6379 redis

# Postgres accepts TCP connections slightly before it's ready to
# authenticate/serve queries during its first-time initialization; give it a
# moment so the app's first connection attempt doesn't predictably fail.
sleep 3

echo "==> Starting application server (this also applies database migrations)"
started=0
for attempt in $(seq 1 5); do
  ./groceries > /tmp/groceries-e2e.log 2>&1 &
  APP_PID=$!

  healthy=0
  for _ in $(seq 1 30); do
    if ! kill -0 "$APP_PID" 2>/dev/null; then
      break
    fi
    if curl --silent --fail --output /dev/null "http://localhost:${PORT:-3000}/login"; then
      healthy=1
      break
    fi
    sleep 1
  done

  if [ "$healthy" = "1" ]; then
    started=1
    break
  fi

  echo "==> Application did not become healthy on attempt ${attempt}, retrying..."
  kill "$APP_PID" 2>/dev/null || true
  wait "$APP_PID" 2>/dev/null || true
  APP_PID=""
  sleep 2
done

if [ "$started" != "1" ]; then
  echo "Application server never became healthy. Logs:" >&2
  cat /tmp/groceries-e2e.log >&2 || true
  exit 1
fi

echo "==> Seeding database with deterministic test data"
mise run seed

echo "==> Installing Playwright and its dependencies"
cd tests
npm ci
npx playwright install --with-deps chromium

echo "==> Running Playwright end-to-end tests"
npx playwright test
