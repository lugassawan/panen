import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/svelte";
import { afterEach } from "vitest";

// Default locale to "en" for tests — all t() calls return English, and
// formatRupiah etc. use en-US locale. Tests that were written with id-ID
// format expectations have their own i18n mock.
Object.defineProperty(navigator, "language", { writable: true, value: "en-US" });

// Persist locale in localStorage so the i18n module picks it up reliably.
try {
  localStorage.setItem("panen-locale", "en");
} catch {
  // localStorage may be unavailable
}

afterEach(() => {
  cleanup();
});
