// k6 load test for the raw /ws pub/sub endpoint.
//
// Each virtual user opens one WebSocket connection, subscribes to all 10 topics,
// and reads "data" frames for the lifetime of the iteration. It measures:
//   - ws_message_latency_ms : Date.now() - msg.sentAt (end-to-end propagation)
//   - ws_messages_received  : total data frames received
//   - ws_messages_missed    : gaps in per-topic `seq` (the silent slow-client drops)
//   - ws_connect_errors     : failures to establish the connection (pre-open errors)
//
// Run (server must be up and streaming a paced session — see loadtest/run.sh):
//   k6 run loadtest/ws_subscribers.js
//   WS_URL=ws://localhost:8080/ws MAX_VUS=500 HOLD=90s k6 run loadtest/ws_subscribers.js

import { WebSocket } from 'k6/experimental/websockets';
import { Trend, Counter } from 'k6/metrics';

const BASE_URL = __ENV.WS_URL || 'ws://localhost:8080/ws';
// Production /ws requires a valid session token (auth.ValidateToken). Browsers can't set
// WebSocket headers, so the token rides as a ?token= query param, matching the real client.
// Locally (APP_ENV=local) auth is skipped, so TOKEN can be left unset.
const TOKEN = __ENV.TOKEN || '';
const WS_URL = TOKEN
  ? `${BASE_URL}${BASE_URL.includes('?') ? '&' : '?'}token=${TOKEN}`
  : BASE_URL;
const MAX_VUS = parseInt(__ENV.MAX_VUS || '400', 10);
const RAMP = __ENV.RAMP || '30s';
const HOLD = __ENV.HOLD || '60s';
// How long each VU keeps a connection open before closing and starting a new iteration.
const CONN_SECONDS = parseInt(__ENV.CONN_SECONDS || '30', 10);

const TOPICS = [
  'lapTimes',
  'positions',
  'carData',
  'raceControl',
  'stints',
  'sectors',
  'segments',
  'trackStatus',
  'qualifyingData',
  'lapCount',
];

const latency = new Trend('ws_message_latency_ms', true);
const received = new Counter('ws_messages_received');
const missed = new Counter('ws_messages_missed');
const connectErrors = new Counter('ws_connect_errors');

export const options = {
  scenarios: {
    subscribers: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: RAMP, target: MAX_VUS },
        { duration: HOLD, target: MAX_VUS },
        { duration: '10s', target: 0 },
      ],
      gracefulStop: '10s',
    },
  },
  thresholds: {
    ws_message_latency_ms: ['p(95)<500'],
    ws_messages_missed: ['count<1'],
    ws_connect_errors: ['count<1'],
  },
};

export default function () {
  const ws = new WebSocket(WS_URL);
  // Per-connection last-seen sequence per topic, to detect gaps (missed messages).
  const lastSeq = {};
  // Whether the socket successfully opened. We only treat a pre-open error as a
  // real connect failure; errors after open are abnormal closes at test teardown
  // (the ramp-down/gracefulStop kills VUs mid-iteration before the clean close
  // below runs), which are not connection failures and shouldn't be counted.
  let opened = false;

  ws.onopen = function () {
    opened = true;
    ws.send(JSON.stringify({ type: 'subscribe', topics: TOPICS }));
    // Hold the connection open, then close to end this iteration cleanly.
    setTimeout(() => ws.close(), CONN_SECONDS * 1000);
  };

  ws.onmessage = function (e) {
    let msg;
    try {
      msg = JSON.parse(e.data);
    } catch (_) {
      return;
    }
    if (msg.type !== 'data') {
      return; // ignore subscribed / recap / session_changed
    }

    received.add(1);
    if (typeof msg.sentAt === 'number' && msg.sentAt > 0) {
      latency.add(Date.now() - msg.sentAt, { topic: msg.topic });
    }

    const topic = msg.topic;
    const seq = msg.seq;
    if (typeof seq === 'number') {
      const prev = lastSeq[topic];
      if (prev !== undefined && seq > prev + 1) {
        missed.add(seq - prev - 1, { topic });
      }
      lastSeq[topic] = seq;
    }
  };

  ws.onerror = function () {
    // Only a failure to establish the connection counts; post-open errors are
    // teardown-time abnormal closes, not connect errors.
    if (!opened) {
      connectErrors.add(1);
    }
  };
}
