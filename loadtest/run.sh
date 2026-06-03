#!/usr/bin/env bash
#
# End-to-end /ws load test orchestration.
#
# Boots DynamoDB Local (free, no license), creates the tables the app needs, starts the
# paced mock SignalR server and the app (APP_ENV=local, pointed at DynamoDB Local), waits
# for readiness, then runs the k6 subscriber load test. Everything is torn down on exit.
#
# S3 is not used: DISABLE_S3_LOGGER silences per-message archival and the one-shot circuit
# download is tolerated (its 1-2 S3 calls fail quietly against the absent endpoint).
#
# Prerequisites: docker, aws CLI, k6, go.
#
# Usage:
#   loadtest/run.sh
#   MAX_VUS=500 HOLD=90s MOCK_REPLAY_DELAY=3ms loadtest/run.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Use 127.0.0.1, NOT localhost: on macOS localhost resolves to IPv6 [::1] but Docker
# Desktop publishes the port on IPv4 only, so localhost:8000 gives "connection refused".
ENDPOINT="http://127.0.0.1:8000"
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_REGION=us-east-1
export AWS_PAGER=""

MOCK_REPLAY_DELAY="${MOCK_REPLAY_DELAY:-5ms}"
DATA_FILE="${DATA_FILE:-australian-fp2-1.txt}"

DDB_CID=""
MOCK_PID=""
APP_PID=""

cleanup() {
  echo "--- cleaning up ---"
  [[ -n "$APP_PID" ]] && kill "$APP_PID" 2>/dev/null || true
  [[ -n "$MOCK_PID" ]] && kill "$MOCK_PID" 2>/dev/null || true
  [[ -n "$DDB_CID" ]] && docker rm -f "$DDB_CID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "--- starting DynamoDB Local ---"
DDB_CID="$(docker run -d -p 127.0.0.1:8000:8000 amazon/dynamodb-local)"
echo "waiting for DynamoDB Local to be ready..."
ddb_ready=false
for _ in $(seq 1 90); do
  if aws --endpoint-url "$ENDPOINT" dynamodb list-tables >/dev/null 2>&1; then ddb_ready=true; break; fi
  sleep 1
done
if [[ "$ddb_ready" != true ]]; then
  echo "ERROR: DynamoDB Local did not become ready at $ENDPOINT" >&2
  docker logs "$DDB_CID" 2>&1 | tail -20 >&2
  exit 1
fi

create_table() {
  local name="$1" pk="$2" pkT="$3" sk="${4:-}" skT="${5:-}"
  local attrs="AttributeName=$pk,AttributeType=$pkT"
  local keys="AttributeName=$pk,KeyType=HASH"
  if [[ -n "$sk" ]]; then
    attrs="$attrs AttributeName=$sk,AttributeType=$skT"
    keys="$keys AttributeName=$sk,KeyType=RANGE"
  fi
  aws --endpoint-url "$ENDPOINT" dynamodb create-table \
    --table-name "$name" \
    --attribute-definitions $attrs \
    --key-schema $keys \
    --billing-mode PAY_PER_REQUEST >/dev/null 2>&1 || true
}

echo "--- creating tables ---"
# Mirrors livetiming/replay_test.go replaySchemas / infra/stack.go.
create_table meetings    MeetingKey N
create_table sessions    MeetingKey N SessionKey   N
create_table drivers     SessionKey N RacingNumber N
create_table positions   SessionKey N RacingNumber N
create_table timings     SessionKey N RacingNumber N
create_table sectors     SessionKey N Reference    S
create_table trackstatus SessionKey N Utc          N
create_table racecontrol SessionKey N Utc          N
create_table stints      SessionKey N RacingNumber N
create_table weather     SessionKey N Utc          N

echo "--- starting mock SignalR server (delay=$MOCK_REPLAY_DELAY) ---"
MOCK_REPLAY_DELAY="$MOCK_REPLAY_DELAY" go run . mockserver "$DATA_FILE" &
MOCK_PID=$!

echo "--- starting app (APP_ENV=local) ---"
# AWS_ENDPOINT_URL_DYNAMODB (service-specific) routes only DynamoDB to the local container;
# S3 is left pointing at real AWS but never used (DISABLE_S3_LOGGER silences per-message
# archival; the one-shot circuit download's 1-2 S3 calls just fail quietly).
APP_ENV=local \
  DISABLE_S3_LOGGER=1 \
  AWS_ENDPOINT_URL_DYNAMODB="$ENDPOINT" \
  AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 \
  go run . &
APP_PID=$!

echo "waiting for app on :8080 ..."
for _ in $(seq 1 60); do
  if curl -fs -o /dev/null "http://127.0.0.1:8080/"; then break; fi
  sleep 1
done

echo "--- running k6 ---"
WS_URL="${WS_URL:-ws://127.0.0.1:8080/ws}" k6 run "$ROOT/loadtest/ws_subscribers.js"
