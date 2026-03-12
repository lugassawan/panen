import { expect, test } from "./fixtures";

test.describe("Sidebar navigation", () => {
  test("renders all nav groups", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    await expect(nav.getByText("Overview")).toBeVisible();
    await expect(nav.getByText("Research")).toBeVisible();
    await expect(nav.getByText("Portfolio", { exact: true }).first()).toBeVisible();
    await expect(nav.getByText("Account")).toBeVisible();
  });

  test("shows Panen brand heading", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { level: 1, name: "Panen" })).toBeVisible();
  });

  test("highlights active nav item with aria-current", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    const dashboardBtn = nav.getByRole("button", { name: "Dashboard" });
    await expect(dashboardBtn).toHaveAttribute("aria-current", "page");

    const lookupBtn = nav.getByRole("button", { name: "Stock Lookup" });
    await lookupBtn.click();

    await expect(lookupBtn).toHaveAttribute("aria-current", "page");
    await expect(dashboardBtn).not.toHaveAttribute("aria-current", "page");
  });

  test("navigates between pages", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });

    await nav.getByRole("button", { name: "Stock Lookup" }).click();
    await expect(page.getByLabel("Stock ticker")).toBeVisible();

    await nav.getByRole("button", { name: "Dashboard" }).click();
    await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
  });
});

test.describe("Dashboard", () => {
  test("is the default page", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
  });
});

test.describe("Page rendering", () => {
  test("Stock Lookup shows ticker input and risk profile", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    await nav.getByRole("button", { name: "Stock Lookup" }).click();

    await expect(page.getByLabel("Stock ticker")).toBeVisible();
    await expect(page.getByLabel("Risk profile")).toBeVisible();
  });

  test("Portfolio page renders", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    // Use exact match to avoid matching the "Portfolio" group label
    await nav.getByRole("button", { name: /^Portfolio$/ }).click();

    // Without backend data, shows onboarding state
    await expect(page.getByRole("heading", { name: "Set Up Your Brokerage" })).toBeVisible();
  });

  test("Settings page shows 4 tabs", async ({ page }) => {
    await page.goto("/");

    // Settings button is outside the nav list
    await page.getByRole("button", { name: "Settings" }).click();

    await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();
    await expect(page.getByRole("tab", { name: "General" })).toBeVisible();
    await expect(page.getByRole("tab", { name: "Data" })).toBeVisible();
    await expect(page.getByRole("tab", { name: "Storage" })).toBeVisible();
    await expect(page.getByRole("tab", { name: "System" })).toBeVisible();
  });

  test("Watchlist page renders", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    await nav.getByRole("button", { name: "Watchlist" }).click();

    await expect(page.getByRole("heading", { name: "Watchlist" })).toBeVisible();
  });

  test("Comparison page renders", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    await nav.getByRole("button", { name: "Compare" }).click();

    await expect(page.getByRole("heading", { name: "Stock Comparison" })).toBeVisible();
  });
});

test.describe("Command Palette", () => {
  test("opens with Cmd/Ctrl+K", async ({ page }) => {
    await page.goto("/");

    await page.keyboard.press("ControlOrMeta+k");
    await expect(page.getByPlaceholder(/search pages/i)).toBeVisible();
  });

  test("shows search input focused", async ({ page }) => {
    await page.goto("/");

    await page.keyboard.press("ControlOrMeta+k");
    const input = page.getByPlaceholder(/search pages/i);
    await expect(input).toBeVisible();
    await expect(input).toBeFocused();
  });

  test("closes with Escape", async ({ page }) => {
    await page.goto("/");

    await page.keyboard.press("ControlOrMeta+k");
    await expect(page.getByPlaceholder(/search pages/i)).toBeVisible();

    await page.keyboard.press("Escape");
    await expect(page.getByPlaceholder(/search pages/i)).not.toBeVisible();
  });
});

test.describe("Settings interaction", () => {
  test("tab switching works", async ({ page }) => {
    await page.goto("/");

    await page.getByRole("button", { name: "Settings" }).click();
    await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();

    await page.getByRole("tab", { name: "Storage" }).click();
    await expect(page.getByRole("tab", { name: "Storage" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
  });

  test("language select has en/id options", async ({ page }) => {
    await page.goto("/");

    await page.getByRole("button", { name: "Settings" }).click();
    await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();

    const langSelect = page.getByLabel("Language");
    await expect(langSelect).toBeVisible();

    const options = langSelect.locator("option");
    await expect(options).toHaveCount(2);
  });

  test("theme toggle is present", async ({ page }) => {
    await page.goto("/");

    await page.getByRole("button", { name: "Settings" }).click();
    await expect(page.getByRole("heading", { name: "Settings" })).toBeVisible();

    await expect(page.getByLabel("Theme")).toBeVisible();
  });
});
