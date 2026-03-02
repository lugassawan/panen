import { expect, test } from "@playwright/test";

test.describe("App navigation", () => {
  test("renders sidebar with navigation items", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    await expect(nav.getByText("Stock Lookup")).toBeVisible();
    await expect(nav.getByText("Portfolio")).toBeVisible();
    await expect(nav.getByText("Settings")).toBeVisible();
  });

  test("shows app name in sidebar", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { level: 1, name: "Panen" })).toBeVisible();
  });

  test("starts on Stock Lookup page by default", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByLabel("Stock ticker")).toBeVisible();
    await expect(page.getByLabel("Risk profile")).toBeVisible();
  });

  test("navigates to Portfolio page", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    await nav.getByText("Portfolio").click();

    await expect(page.getByText("Portfolio management — coming soon")).toBeVisible();
    await expect(page.getByLabel("Stock ticker")).not.toBeVisible();
  });

  test("navigates to Settings page", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    await nav.getByText("Settings").click();

    await expect(page.getByLabel("Language")).toBeVisible();
    await expect(page.getByLabel("Theme")).toBeVisible();
    await expect(page.getByText("Coming in a future update")).toBeVisible();
  });

  test("returns to Stock Lookup from another page", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    await nav.getByText("Portfolio").click();
    await expect(page.getByText("Portfolio management — coming soon")).toBeVisible();

    await nav.getByText("Stock Lookup").click();
    await expect(page.getByLabel("Stock ticker")).toBeVisible();
  });

  test("highlights active nav item", async ({ page }) => {
    await page.goto("/");

    const nav = page.getByRole("navigation", { name: /main/i });
    const lookupBtn = nav.getByText("Stock Lookup").locator("..");
    const portfolioBtn = nav.getByText("Portfolio").locator("..");

    await expect(lookupBtn).toHaveAttribute("aria-current", "page");
    await expect(portfolioBtn).not.toHaveAttribute("aria-current");

    await portfolioBtn.click();

    await expect(portfolioBtn).toHaveAttribute("aria-current", "page");
    await expect(lookupBtn).not.toHaveAttribute("aria-current");
  });
});
