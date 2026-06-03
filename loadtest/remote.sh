#!/usr/bin/env bash
#
# Load-test a REMOTE /ws endpoint (e.g. production wss://pushlap.co/ws) with k6,
# optionally sampling the server container's CPU/RAM over SSH for the duration.
#
# Unlike run.sh (which boots the whole stack locally), this ONLY runs the k6
# generator against an already-running server. Always run k6 from a DIFFERENT
# machine than the server, so the generator never competes with it for CPU.
#
# Usage:
#   loadtest/remote.sh
#   WS_URL=wss://pushlap.co/ws MAX_VUS=50 loadtest/remote.sh
#   SSH_TARGET=ec2-user@<ec2-ip> loadtest/remote.sh   # also captures server docker stats
#
# CAUTION: this puts real load on a live service. Pick a quiet window and start
# with a low MAX_VUS. latency/missed only register while the server is actively
# broadcasting (position replay is continuous; other topics need a live session).

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

WS_URL="${WS_URL:-wss://pushlap.co/ws}"
# Set SSH_TARGET (e.g. ec2-user@1.2.3.4) to sample the server container during the
# run. Requires key-based SSH access to the box. Leave unset to skip (use CloudWatch).
SSH_TARGET="${SSH_TARGET:-}"
# Identity file for SSH (the EC2 key pair). Defaults to infra/f1telemetry.pem if present.
SSH_KEY="${SSH_KEY:-$ROOT/infra/f1telemetry.pem}"
CONTAINER="${CONTAINER:-f1telemetry}"
STATS_INTERVAL="${STATS_INTERVAL:-5}"

# Build the ssh argument list, adding -i only when the key file exists.
SSH_ARGS=(-o BatchMode=yes -o ConnectTimeout=5)
[[ -f "$SSH_KEY" ]] && SSH_ARGS+=(-i "$SSH_KEY")

STATS_LOG=""
STATS_PID=""

cleanup() {
  [[ -n "$STATS_PID" ]] && kill "$STATS_PID" 2>/dev/null || true
}
trap cleanup EXIT

if [[ -n "$SSH_TARGET" ]]; then
  STATS_LOG="$(mktemp -t f1-docker-stats.XXXXXX)"
  echo "--- sampling '$CONTAINER' on $SSH_TARGET every ${STATS_INTERVAL}s ---"
  (
    while true; do
      ts="$(date +%H:%M:%S)"
      line="$(ssh "${SSH_ARGS[@]}" "$SSH_TARGET" \
        "docker stats --no-stream --format '{{.CPUPerc}} {{.MemUsage}}' $CONTAINER" 2>/dev/null || echo 'n/a')"
      echo "$ts $line" >>"$STATS_LOG"
      sleep "$STATS_INTERVAL"
    done
  ) &
  STATS_PID=$!
fi

echo "--- running k6 against $WS_URL ---"
WS_URL="$WS_URL" k6 run "$ROOT/loadtest/ws_subscribers.js"

if [[ -n "$STATS_LOG" ]]; then
  echo "--- server container samples ($CONTAINER on $SSH_TARGET) ---"
  cat "$STATS_LOG"
  # Peak CPU% across samples ($2 is the CPUPerc column, e.g. "73.20%").
  awk '{gsub(/%/,"",$2); if($2+0>m){m=$2+0}} END{if(NR)print "peak container CPU: "m"%"}' "$STATS_LOG"
else
  echo "--- (no SSH_TARGET set: check the instance's CloudWatch CPUUtilization for server load) ---"
fi
