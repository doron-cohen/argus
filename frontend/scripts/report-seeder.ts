#!/usr/bin/env bun

const DEFAULT_BASE_URL = "http://localhost:8080";
const REPORTS_API_BASE = "/api/reports/v1";
const CATALOG_API_BASE = "/api/catalog/v1";

type SeedOptions = {
  excludeComponents?: string[];
  includeOnly?: string[] | null;
  reportsPerComponent?: number;
  includeAllStatuses?: boolean;
  verbose?: boolean;
};

function log(msg: string) {
  console.log(msg);
}

const CHECK_TYPES = [
  {
    slug: "unit-tests",
    name: "Unit Tests",
    description: "Runs unit tests for the component",
    successRate: 0.8,
  },
  {
    slug: "security-scan",
    name: "Security Scan",
    description: "Performs security vulnerability scanning",
    successRate: 0.7,
  },
  {
    slug: "code-quality",
    name: "Code Quality",
    description: "Analyzes code quality metrics",
    successRate: 0.6,
  },
  {
    slug: "build",
    name: "Build",
    description: "Compiles and builds the component",
    successRate: 0.9,
  },
  {
    slug: "integration-tests",
    name: "Integration Tests",
    description: "Runs integration tests",
    successRate: 0.75,
  },
];

const STATUSES = [
  "pass",
  "fail",
  "error",
  "disabled",
  "skipped",
  "unknown",
  "completed",
];

export class ReportSeeder {
  constructor(private readonly baseUrl: string = DEFAULT_BASE_URL) {}

  private api(path: string): string {
    return `${this.baseUrl.replace(/\/$/, "")}${path}`;
  }

  async getComponents(): Promise<Array<{ id?: string; name?: string }>> {
    log("🔍 Fetching components from API...");
    const response = await fetch(this.api(`${CATALOG_API_BASE}/components`));
    if (!response.ok)
      throw new Error(
        `Failed to fetch components: ${response.status} ${response.statusText}`
      );
    const components = (await response.json()) as Array<{
      id?: string;
      name?: string;
    }>;
    log(`✅ Found ${components.length} components`);
    return components;
  }

  async submitReport(report: any): Promise<any> {
    const response = await fetch(this.api(`${REPORTS_API_BASE}/reports`), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(report),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(
        `Failed to submit report: ${response.status} ${response.statusText} - ${errorText}`
      );
    }
    return response.json();
  }

  private selectStatusForCheck(
    checkType: (typeof CHECK_TYPES)[number]
  ): (typeof STATUSES)[number] {
    const random = Math.random();
    if (random < checkType.successRate) return "pass";
    const remaining = 1 - checkType.successRate;
    const failThreshold = checkType.successRate + remaining * 0.5;
    const errorThreshold = failThreshold + remaining * 0.2;
    if (random < failThreshold) return "fail";
    if (random < errorThreshold) return "error";
    const other = ["disabled", "skipped", "unknown", "completed"] as const;
    return other[Math.floor(Math.random() * other.length)];
  }

  private generateReportData(
    componentId: string,
    checkType: (typeof CHECK_TYPES)[number],
    status: string
  ) {
    const ts = new Date();
    ts.setMinutes(ts.getMinutes() - Math.floor(Math.random() * 60));
    const report: any = {
      check: {
        slug: checkType.slug,
        name: checkType.name,
        description: checkType.description,
      },
      component_id: componentId,
      status,
      timestamp: ts.toISOString(),
      metadata: {
        ci_job_id: `job-${Math.random().toString(36).slice(2, 11)}`,
        environment: Math.random() > 0.5 ? "staging" : "production",
        branch: Math.random() > 0.8 ? "feature/test" : "main",
        commit_sha: Math.random().toString(36).slice(2, 9),
        execution_duration_ms: 60000,
      },
    };
    return report;
  }

  async seedReportsForComponent(
    componentId: string,
    options: {
      checksPerComponent: number;
      includeAllStatuses: boolean;
      verbose: boolean;
    }
  ) {
    if (options.verbose)
      log(`📊 Creating reports for component: ${componentId}`);
    const selectedChecks = CHECK_TYPES.slice(0, options.checksPerComponent);
    const created: Array<{
      checkSlug: string;
      status: string;
      reportId: string;
    }> = [];
    for (const check of selectedChecks) {
      const status = options.includeAllStatuses
        ? STATUSES[selectedChecks.indexOf(check) % STATUSES.length]
        : this.selectStatusForCheck(check);
      const report = this.generateReportData(componentId, check, status);
      const result = await this.submitReport(report);
      created.push({
        checkSlug: check.slug,
        status,
        reportId: result.report_id,
      });
      if (options.verbose) log(`  ✅ ${check.slug}: ${status}`);
    }
    return created;
  }

  async seedAll(options: SeedOptions = {}) {
    const {
      excludeComponents = [],
      includeOnly = null,
      reportsPerComponent = 3,
      includeAllStatuses = false,
      verbose = true,
    } = options;
    log("🌱 Starting report seeding...");
    const components = await this.getComponents();
    let targets = components;
    if (includeOnly && includeOnly.length)
      targets = components.filter((c) =>
        includeOnly.includes((c.id || c.name) as string)
      );
    else if (excludeComponents.length)
      targets = components.filter(
        (c) => !excludeComponents.includes((c.id || c.name) as string)
      );
    log(`🎯 Targeting ${targets.length} components for seeding`);
    const all: Record<string, any> = {};
    for (const c of targets) {
      const id = (c.id || c.name) as string;
      all[id] = await this.seedReportsForComponent(id, {
        checksPerComponent: reportsPerComponent,
        includeAllStatuses,
        verbose,
      });
    }
    log("🎉 Seeding completed successfully!");
    return all;
  }
}

// CLI usage (optional)
if (import.meta.main) {
  (async () => {
    const args = Bun.argv.slice(2);
    const getFlag = (name: string) => args.includes(name);
    const getValue = (name: string, def?: string) => {
      const i = args.indexOf(name);
      return i >= 0 && i + 1 < args.length ? args[i + 1] : def;
    };

    const baseUrl = getValue(
      "--base-url",
      Bun.env.ARGUS_BASE_URL || DEFAULT_BASE_URL
    ) as string;

    const onlyList: string[] = [];
    for (let i = 0; i < args.length; i++) {
      if (args[i] === "--only" && i + 1 < args.length) onlyList.push(args[i + 1]);
    }
    const excludeList: string[] = [];
    for (let i = 0; i < args.length; i++) {
      if (args[i] === "--exclude" && i + 1 < args.length)
        excludeList.push(args[i + 1]);
    }
    const reportsPer = parseInt(
      getValue("--reports-per-component", "3") as string,
      10
    );
    const includeAllStatuses = getFlag("--all-statuses");
    const verbose = !getFlag("--quiet");

    const seeder = new ReportSeeder(baseUrl);
    await seeder.seedAll({
      includeOnly: onlyList.length ? onlyList : null,
      excludeComponents: excludeList,
      reportsPerComponent: reportsPer,
      includeAllStatuses,
      verbose,
    });
  })().catch((err) => {
    console.error(err?.message || err);
    process.exit(1);
  });
}


