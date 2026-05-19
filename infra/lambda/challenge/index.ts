import { randomUUID } from "crypto";
import Redis from "ioredis";

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

export const handler = async () => {
  const id = randomUUID();
  await getRedis().set(`challenge:${id}`, "pending", "EX", 300);
  return {
    statusCode: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ challenge: id }),
  };
};
