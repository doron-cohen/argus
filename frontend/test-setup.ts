import { GlobalRegistrator } from "@happy-dom/global-registrator";
import { beforeEach, afterEach } from "bun:test";

// Register DOM globals before tests start
GlobalRegistrator.register();

// Clean up DOM after each test
afterEach(() => {
  if (typeof document !== "undefined") {
    document.body.innerHTML = "";
  }
});
