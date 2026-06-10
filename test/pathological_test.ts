import { createClient, type RedisClientType } from "redis";
import { assertEquals } from "@std/assert";

interface TestCase {
  name: string;
  command: string[];
  expected: unknown;
  type: string;
  version: string;
  description?: string;
  setup?: string[][];
}

interface TestResult {
  name: string;
  version: string;
  status: "passed" | "failed" | "skipped";
  expected: unknown;
  actual: unknown;
  error?: string;
}

const testCases: TestCase[] = JSON.parse(
  new TextDecoder().decode(Deno.readFileSync("pathological_test.json"))
);

let client: RedisClientType | null = null;
const results: TestResult[] = [];

function compareResult(actual: unknown, expected: unknown, type: string): boolean {
  switch (type) {
    case "simple_string":
    case "bulk_string":
      return actual === expected;
    case "integer":
      return Number(actual) === Number(expected);
    case "null":
      return actual === null;
    case "array":
      if (!Array.isArray(expected) || !Array.isArray(actual)) return false;
      if (expected.length !== actual.length) return false;
      for (let i = 0; i < expected.length; i++) {
        if (expected[i] === null) {
          if (actual[i] !== null) return false;
        } else {
          if (actual[i] !== expected[i]) return false;
        }
      }
      return true;
    default:
      return actual === expected;
  }
}

function formatResult(actual: unknown): string {
  if (actual === null) return "null";
  if (actual === undefined) return "undefined";
  if (Array.isArray(actual)) {
    return "[" + actual.map(formatResult).join(", ") + "]";
  }
  return String(actual);
}

async function executeCommand(client: RedisClientType, cmd: string[]): Promise<unknown> {
  const command = cmd[0].toLowerCase();
  const args = cmd.slice(1);

  try {
    switch (command) {
      case "ping":
        if (args.length > 0) {
          return await client.ping(args[0]);
        }
        return await client.ping();
      case "set":
        await client.set(args[0], args[1]);
        return "OK";
      case "get":
        return await client.get(args[0]);
      case "del":
        const delResult = await client.del(args);
        return delResult;
      default:
        try {
          const result = await client.sendCommand([cmd[0], ...args]);
          return result;
        } catch (err) {
          return { error: String(err) };
        }
    }
  } catch (err) {
    return { error: String(err) };
  }
}

async function runTest(testCase: TestCase): Promise<TestResult> {
  try {
    if (!client) throw new Error("Client not connected");

    if (testCase.setup) {
      for (const setupCmd of testCase.setup) {
        await executeCommand(client, setupCmd);
      }
    }

    const actual = await executeCommand(client, testCase.command);

    if (actual && typeof actual === "object" && "error" in actual) {
      return {
        name: testCase.name,
        version: testCase.version,
        status: "failed",
        expected: testCase.expected,
        actual: null,
        error: (actual as { error: string }).error,
      };
    }

    const passed = compareResult(actual, testCase.expected, testCase.type);

    return {
      name: testCase.name,
      version: testCase.version,
      status: passed ? "passed" : "failed",
      expected: testCase.expected,
      actual: actual,
      error: passed ? undefined : `Expected ${formatResult(testCase.expected)}, got ${formatResult(actual)}`,
    };
  } catch (err) {
    return {
      name: testCase.name,
      version: testCase.version,
      status: "failed",
      expected: testCase.expected,
      actual: null,
      error: String(err),
    };
  }
}

function printResultsTable(results: TestResult[]) {
  console.log("\n" + "=".repeat(80));
  console.log("PATHOLOGICAL TEST RESULTS");
  console.log("=".repeat(80));

  const passed = results.filter(r => r.status === "passed").length;
  const failed = results.filter(r => r.status === "failed").length;

  console.log(`Total: ${results.length} | Passed: ${passed} | Failed: ${failed}`);
  console.log("-".repeat(80));

  for (const result of results) {
    const statusStr = result.status === "passed" ? "PASS" : "FAIL";
    const line = `${statusStr} | ${result.version} | ${result.name}`;
    console.log(line);

    if (result.error) {
      console.log(`  Error: ${result.error}`);
    }
  }

  console.log("=".repeat(80));
  console.log(`\nSummary: ${passed}/${results.length} tests passed`);

  if (failed > 0) {
    console.log("\nFailed tests:");
    for (const result of results.filter(r => r.status === "failed")) {
      console.log(`  - ${result.name} (v${result.version})`);
      if (result.error) {
        console.log(`    ${result.error}`);
      }
    }
  }
}

Deno.test("Pathological JSON Command Test Suite", async () => {
  client = await createClient({
    url: "redis://localhost:6379",
    socket: {
      reconnectStrategy: false,
    },
  });

  try {
    await client.connect();

    for (const testCase of testCases) {
      const result = await runTest(testCase);
      results.push(result);

      if (result.status === "failed") {
        console.error(`FAILED: ${result.name}`);
        if (result.error) {
          console.error(`  ${result.error}`);
        }
      }
    }

    printResultsTable(results);

    const failedResults = results.filter(r => r.status === "failed");
    assertEquals(
      failedResults.length,
      0,
      `Some tests failed: ${failedResults.map(r => r.name).join(", ")}`
    );

  } finally {
    if (client) {
      await client.quit();
    }
  }
});
