import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

// Mock the theme store to avoid localStorage initialization issues.
let mockPreference = "light";
vi.mock("../stores/theme.svelte", () => ({
  theme: {
    get preference() {
      return mockPreference;
    },
    toggle() {
      const cycle = ["light", "dark", "system"];
      mockPreference = cycle[(cycle.indexOf(mockPreference) + 1) % cycle.length];
    },
    set(pref: string) {
      mockPreference = pref;
    },
  },
}));

import ThemeToggle from "./ThemeToggle.svelte";

describe("ThemeToggle", () => {
  beforeEach(() => {
    mockPreference = "light";
  });

  it("renders toggle button with aria-label", () => {
    render(ThemeToggle);
    expect(screen.getByRole("button")).toHaveAttribute("aria-label", "Toggle theme (light)");
  });

  it("cycles theme on click", async () => {
    const user = userEvent.setup();
    render(ThemeToggle);

    // light → dark
    await user.click(screen.getByRole("button"));
    expect(mockPreference).toBe("dark");

    // dark → system
    await user.click(screen.getByRole("button"));
    expect(mockPreference).toBe("system");

    // system → light
    await user.click(screen.getByRole("button"));
    expect(mockPreference).toBe("light");
  });
});
