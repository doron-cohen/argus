import { playwrightLauncher } from "@web/test-runner-playwright";
import { esbuildPlugin } from "@web/dev-server-esbuild";

export default {
  files: ["src/**/*.test.ts", "tests/unit/**/*.test.ts"],
  nodeResolve: true,
  browsers: [playwrightLauncher({ product: "chromium" })],
  testsFinishTimeout: 30000,
  plugins: [
    esbuildPlugin({
      ts: true,
      target: "es2020",
      loaders: { ".ts": "ts" },
      tsconfig: "tsconfig.json",
      define: { "process.env.NODE_ENV": '"test"' },
    }),
  ],
};
