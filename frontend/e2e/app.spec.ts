import { expect, test } from "@playwright/test";

test.describe("App", () => {
  test("renders the heading and description", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByRole("heading", { level: 1 })).toHaveText("Panen");
    await expect(page.getByText("Desktop decision engine for IDX investors")).toBeVisible();
  });

  test("renders input and greet button", async ({ page }) => {
    await page.goto("/");

    await expect(page.getByPlaceholder("Enter your name")).toBeVisible();
    await expect(page.getByRole("button", { name: "Greet" })).toBeVisible();
  });
});
