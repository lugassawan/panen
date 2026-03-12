import { test as base } from "@playwright/test";

/**
 * Extended Playwright test fixture that stubs the Wails runtime before each page load.
 *
 * E2E tests run against the standalone Vite dev server (port 5173) where the
 * Wails runtime is absent.  Without these stubs the app crashes on mount
 * because store modules (sync, update, alerts) call `window.runtime.EventsOnMultiple`
 * at import time, and page components call `window.go.backend.App.*` methods.
 */
export const test = base.extend({
  page: async ({ page }, use) => {
    await page.addInitScript(() => {
      // Stub window.runtime — used by wailsjs/runtime/runtime.js
      // Every method is a no-op that returns a no-op (e.g. EventsOnMultiple returns a cancel fn)
      (window as Record<string, unknown>).runtime = new Proxy(
        {},
        {
          get() {
            return () => () => {};
          },
        },
      );

      // Stub window.go.backend.App — used by wailsjs/go/backend/App.js
      // Each method returns a resolved promise with null.
      (window as Record<string, unknown>).go = {
        backend: {
          App: new Proxy(
            {},
            {
              get() {
                return (..._args: unknown[]) => Promise.resolve(null);
              },
            },
          ),
        },
      };
    });

    await use(page);
  },
});

export { expect } from "@playwright/test";
