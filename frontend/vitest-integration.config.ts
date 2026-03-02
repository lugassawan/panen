import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [svelte()],
  resolve: {
    conditions: ["browser"],
  },
  test: {
    environment: "jsdom",
    include: ["src/**/*.integration.test.ts"],
    passWithNoTests: true,
    testTimeout: 10_000,
    setupFiles: ["src/test-setup.ts"],
    coverage: {
      provider: "v8",
      reportsDirectory: "../coverage/frontend-integration",
      include: ["src/**/*.{ts,svelte}"],
      exclude: ["src/test-setup.ts", "src/vite-env.d.ts"],
    },
  },
});
