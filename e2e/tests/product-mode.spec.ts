import { spawn, type ChildProcess } from "node:child_process";
import { existsSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import {
  test,
  expect,
  buildBinary,
  initGitRepo,
  ralphBinaryPath,
  startServerInWorkDir,
} from "../helpers/server.ts";

test("loads a product document in the local run view", async ({ page }) => {
  buildBinary();

  const workDir = mkdtempSync(join(tmpdir(), "ralph-product-e2e-"));
  initGitRepo(workDir);
  writeFileSync(join(workDir, "main.go"), "package main\n");
  writeFileSync(
    join(workDir, "product.json"),
    JSON.stringify(
      {
        version: 1,
        project_name: "Product Mode",
        stories: [
          {
            id: "story-1",
            title: "Outcomes",
            description: "Focus on value",
            slices: [
              {
                id: "slice-1",
                behavior: "ship the outcome",
                passes: false,
              },
            ],
            priority: 1,
            passes: false,
          },
        ],
      },
      null,
      2,
    ),
  );

  const server = await startServerInWorkDir(workDir);
  try {
    await page.goto(`${server.baseURL}/runs/prd-local`);

    await expect(page).toHaveURL(/\/runs\/prd-local$/);
    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Completed");
    await expect(page.getByText("Product Mode")).toBeVisible();
    await expect(page.getByText("This run is in progress in the terminal")).toBeVisible();
  } finally {
    server.stop();
  }
});

test("--product generates product.json without implementation fields", async () => {
  buildBinary();

  const workDir = mkdtempSync(join(tmpdir(), "ralph-product-cli-e2e-"));
  initGitRepo(workDir);
  writeFileSync(join(workDir, "main.go"), "package main\n");

  const cli = spawnRalphProduct(workDir);
  try {
    await waitForProduct(workDir);
    const raw = readFileSync(join(workDir, "product.json"), "utf8");
    const product = JSON.parse(raw) as Record<string, unknown>;

    expect(product.project_name).toBe("Mock Product");
    expect(raw).not.toContain("red_hint");
    expect(raw).not.toContain("context");
    expect(raw).not.toContain("test_spec");
    expect(existsSync(join(workDir, "prd.json"))).toBe(false);
  } finally {
    cli.kill("SIGTERM");
  }
});

function spawnRalphProduct(workDir: string): ChildProcess {
  const bin = ralphBinaryPath();
  const args =
    process.platform === "darwin"
      ? ["-q", "/dev/null", bin, "--product", "build a widget"]
      : ["-q", "-c", `${shellQuote(bin)} --product ${shellQuote("build a widget")}`, "/dev/null"];

  return spawn("script", args, {
    cwd: workDir,
    env: {
      ...process.env,
      RALPH_RUNNER: "mock",
    },
    stdio: "ignore",
  });
}

function shellQuote(value: string): string {
  return `'${value.replaceAll("'", `'\\''`)}'`;
}

async function waitForProduct(workDir: string): Promise<void> {
  await waitUntil(() => existsSync(join(workDir, "product.json")), "product.json to exist");
}

async function waitUntil(check: () => boolean, label: string): Promise<void> {
  const deadline = Date.now() + 90_000;
  while (Date.now() < deadline) {
    if (check()) return;
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`timed out waiting for ${label}`);
}
