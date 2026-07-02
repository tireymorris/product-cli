import { writeFileSync } from "node:fs";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { test, expect, buildBinary, initGitRepo, startServerInWorkDir } from "../helpers/server.ts";

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
    await expect(page.locator(".app-main .run-status-badge")).toHaveText("Implementing");
    await expect(page.locator(".run-detail-progress-label")).toHaveText("0/1");
    await expect(page.getByText("Product Mode")).toBeVisible();
    await expect(page.getByText("This run is in progress in the terminal")).toBeVisible();
  } finally {
    server.stop();
  }
});
