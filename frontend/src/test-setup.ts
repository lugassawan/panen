import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/svelte";
import { afterEach } from "vitest";

// Default test locale to "id" to match original id-ID number formatting behavior.
// Individual tests can override navigator.language as needed.
Object.defineProperty(navigator, "language", { writable: true, value: "id-ID" });

afterEach(() => {
  cleanup();
});
