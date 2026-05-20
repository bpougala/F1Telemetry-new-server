import { randomUUID, createHash } from "crypto";
import Redis from "ioredis";
// @ts-ignore — no type declarations
import { verifyAttestation, verifyAssertion } from "appattest-checker-node";

const SESSION_TTL = 604800; // 7 days in seconds

let redis: Redis | null = null;

function getRedis(): Redis {
  if (!redis) {
    redis = new Redis({
      host: process.env.VALKEY_ENDPOINT!,
      port: parseInt(process.env.VALKEY_PORT || "6379", 10),
      tls: {},
    });
  }
  return redis;
}

function response(statusCode: number, body: object) {
  return {
    statusCode,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  };
}

async function issueSessionToken(keyId: string): Promise<string> {
  const token = randomUUID();
  await getRedis().set(`session:${token}`, keyId, "EX", SESSION_TTL);
  return token;
}

async function handleAttestation(body: {
  keyId: string;
  attestation: string;
  challenge: string;
}) {
  const { keyId, attestation, challenge } = body;
  if (!keyId || !attestation || !challenge) {
    return response(400, { error: "missing keyId, attestation, or challenge" });
  }

  const client = getRedis();

  // Verify and consume the challenge (one-time use)
  const challengeVal = await client.get(`challenge:${challenge}`);
  if (challengeVal !== "pending") {
    return response(401, { error: "invalid or expired challenge" });
  }
  await client.del(`challenge:${challenge}`);

  // Verify attestation with Apple
  const appInfo = {
    appId: process.env.APP_ID!,
    develomentEnv: process.env.DEV_ENV === "true",
  };

  const attestationBuffer = Buffer.from(attestation, "base64");
  const result = await verifyAttestation(
    appInfo,
    keyId,
    challenge,
    attestationBuffer
  );

  if (result.verifyError) {
    console.error("attestation verifyError:", JSON.stringify(result.verifyError));
    return response(401, { error: "attestation verification failed" });
  }

  // Store the device's public key permanently
  await client.set(
    `key:${keyId}`,
    JSON.stringify({ publicKeyPem: result.publicKeyPem, signCount: 0 })
  );

  const token = await issueSessionToken(keyId);
  return response(200, { token });
}

async function handleAssertion(body: {
  keyId: string;
  assertion: string;
  challenge: string;
}) {
  const { keyId, assertion, challenge } = body;
  if (!keyId || !assertion || !challenge) {
    return response(400, { error: "missing keyId, assertion, or challenge" });
  }

  const client = getRedis();

  // Verify and consume the challenge
  const challengeVal = await client.get(`challenge:${challenge}`);
  if (challengeVal !== "pending") {
    return response(401, { error: "invalid or expired challenge" });
  }
  await client.del(`challenge:${challenge}`);

  // Retrieve stored key data
  const keyData = await client.get(`key:${keyId}`);
  if (!keyData) {
    return response(401, { error: "unknown device key" });
  }

  const { publicKeyPem, signCount } = JSON.parse(keyData);

  // Compute clientDataHash = SHA256(challenge)
  const clientDataHash = createHash("sha256")
    .update(challenge)
    .digest();

  const assertionBuffer = Buffer.from(assertion, "base64");
  const result = await verifyAssertion(
    clientDataHash,
    publicKeyPem,
    process.env.APP_ID!,
    assertionBuffer
  );

  if (result.verifyError) {
    return response(401, { error: "assertion verification failed" });
  }

  // Verify sign count increased (replay protection)
  if (result.signCount <= signCount) {
    return response(401, { error: "replay detected" });
  }

  // Update stored sign count
  await client.set(
    `key:${keyId}`,
    JSON.stringify({ publicKeyPem, signCount: result.signCount })
  );

  const token = await issueSessionToken(keyId);
  return response(200, { token });
}

async function handleDev(body: { token: string }) {
  const devToken = process.env.DEV_TOKEN;
  if (!devToken) {
    return response(403, { error: "dev bypass disabled" });
  }
  if (body.token !== devToken) {
    return response(403, { error: "invalid dev token" });
  }

  const token = await issueSessionToken("dev");
  return response(200, { token });
}

export const handler = async (event: any) => {
  try {
    const body = JSON.parse(event.body || "{}");

    switch (body.type) {
      case "attest":
        return handleAttestation(body);
      case "assert":
        return handleAssertion(body);
      case "dev":
        return handleDev(body);
      default:
        return response(400, { error: "invalid type, expected: attest|assert|dev" });
    }
  } catch (err) {
    console.error("handler error:", err);
    return response(500, { error: "internal server error" });
  }
};
