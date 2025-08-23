import { playwrightLauncher } from "@web/test-runner-playwright";
import { esbuildPlugin } from "@web/dev-server-esbuild";

export default {
  files: "src/**/*.test.ts",
  nodeResolve: true,
  browsers: [playwrightLauncher({ product: "chromium" })],
  plugins: [
    esbuildPlugin({
      ts: true,
      target: "es2020",
      loaders: { ".ts": "ts" },
      tsconfig: "tsconfig.json",
    }),
  ],
};
