import type { APIRequestContext } from "@playwright/test";

const BASE_URL = process.env.ARGUS_BASE_URL || "http://localhost:8080";

export async function waitForSync(
  request: APIRequestContext,
  timeoutMs = 30000,
): Promise<void> {
  const start = Date.now();
  // Poll sync status until completed/success or timeout
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await request.get(`${BASE_URL}/api/sync/v1/sources/0/status`);
      if (res.ok()) {
        const data = await res.json();
        if (
          data.status === "completed" ||
          data.lastSync?.status === "success"
        ) {
          return;
        }
      }
    } catch {
      // ignore and retry
    }
    await new Promise((r) => setTimeout(r, 500));
  }
  throw new Error("Timed out waiting for sync to complete");
}

export type CreateReportOptions = {
  checkSlug?: string;
  checkName?: string;
  checkDescription?: string;
  status?:
    | "pass"
    | "fail"
    | "disabled"
    | "skipped"
    | "unknown"
    | "error"
    | "completed";
  timestampIso?: string;
};

export async function createReport(
  request: APIRequestContext,
  componentId: string,
  options: CreateReportOptions = {},
): Promise<{ report_id: string } | undefined> {
  const nowIso = new Date().toISOString();
  const body = {
    check: {
      slug:
        options.checkSlug || `e2e-${Math.random().toString(36).slice(2, 10)}`,
      name: options.checkName || "E2E Fixture",
      description: options.checkDescription || "Created by Playwright fixture",
    },
    component_id: componentId,
    status: options.status || "pass",
    timestamp: options.timestampIso || nowIso,
    metadata: {
      ci_job_id: `job-${Math.random().toString(36).slice(2, 11)}`,
      environment: "test",
      branch: "e2e",
      commit_sha: Math.random().toString(36).slice(2, 9),
      execution_duration_ms: 12345,
    },
  };

  const res = await request.post(`${BASE_URL}/api/reports/v1/reports`, {
    data: body,
  });
  if (!res.ok()) {
    const text = await res.text();
    throw new Error(`Failed to create report: ${res.status()} ${text}`);
  }
  return res.json();
}

export async function getLatestReports(
  request: APIRequestContext,
  componentId: string,
): Promise<{ reports: any[] }> {
  const res = await request.get(
    `${BASE_URL}/api/catalog/v1/components/${componentId}/reports?latest_per_check=true`,
  );
  if (res.status() === 404) return { reports: [] };
  if (!res.ok()) throw new Error(`Failed fetching reports: ${res.status()}`);
  return res.json();
}

export async function ensureReports(
  request: APIRequestContext,
  componentId: string,
  desiredCount = 1,
): Promise<void> {
  try {
    await waitForSync(request, 10000);
  } catch {
    // continue; report creation does not require sync to be completed
  }
  const current = await getLatestReports(request, componentId);
  const toCreate = Math.max(0, desiredCount - current.reports.length);
  for (let i = 0; i < toCreate; i++) {
    await createReport(request, componentId, {
      status: i % 2 ? "fail" : "pass",
    });
  }
}
