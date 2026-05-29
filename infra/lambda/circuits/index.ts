import { S3Client, GetObjectCommand } from "@aws-sdk/client-s3";

const s3 = new S3Client({});
const BUCKET = process.env.CIRCUIT_BUCKET!;

function response(statusCode: number, body: string) {
  return {
    statusCode,
    headers: {
      "Content-Type": "application/json",
      "Cache-Control": "public, max-age=86400",
    },
    body,
  };
}

export const handler = async (event: any) => {
  const params = event.queryStringParameters || {};
  const { circuitKey, year } = params;

  if (!circuitKey || !year) {
    return response(400, JSON.stringify({ error: "missing circuitKey or year query parameter" }));
  }

  if (isNaN(Number(circuitKey)) || isNaN(Number(year))) {
    return response(400, JSON.stringify({ error: "circuitKey and year must be numbers" }));
  }

  try {
    const result = await s3.send(
      new GetObjectCommand({
        Bucket: BUCKET,
        Key: `circuits/${circuitKey}/${year}.json`,
      })
    );

    const body = await result.Body!.transformToString("utf-8");
    return response(200, body);
  } catch (err: any) {
    if (err.name === "NoSuchKey" || err.$metadata?.httpStatusCode === 404) {
      return response(404, JSON.stringify({ error: "circuit data not found" }));
    }
    console.error("handler error:", err);
    return response(500, JSON.stringify({ error: "internal server error" }));
  }
};
