// @ts-check
const { test, expect } = require("@playwright/test");

// These tests rely on the deterministic seed data loaded by `mise run seed`
// (see internal/db/seeds). "Apples" is seeded as an item that exists in the
// catalog but starts off the shopping list, making it a safe item to
// exercise the add/done/checkout flow without disturbing other seeded data.
const TEST_USERNAME = "admin";
const TEST_PASSWORD = "marbleslyra";
const TEST_ITEM = "Apples";

test.describe.configure({ mode: "serial" });

test("shopping list journey: login, add, mark done, and check out", async ({
  page,
}) => {
  await test.step("can log in", async () => {
    await page.goto("/login");

    await page.getByPlaceholder("Username").fill(TEST_USERNAME);
    await page.getByPlaceholder("Password").fill(TEST_PASSWORD);
    await page.getByRole("button", { name: "Login" }).click();

    await expect(page).toHaveURL("/");
    await expect(page.getByRole("link", { name: "Logout" })).toBeVisible();
  });

  await test.step("can dynamically add an existing item to the list", async () => {
    await page.goto("/");

    // The "name" field is backed by a <datalist> of items that are not yet
    // on the shopping list, confirming this exercises the "existing item"
    // path (as opposed to creating a brand new item).
    await expect(
      page.locator('#list-add-items option[value="Apples"]'),
    ).toHaveCount(1);

    await page.locator("#name").fill(TEST_ITEM);
    // Scoped to the form's submit button specifically: BeerCSS renders an
    // icon ligature ahead of the button text (e.g. "add Add"), which makes
    // a plain accessible-name match for "Add" ambiguous with the circle
    // "add_shopping_cart" icon buttons elsewhere on the page.
    await page.locator('button[form="itemAdderForm"]').click();

    await expect(page).toHaveURL("/");
    await expect(page.locator("#list").getByText(TEST_ITEM)).toBeVisible();
  });

  await test.step("can mark an item done in the list and see it appear in the shopping cart", async () => {
    const listItem = page.locator("#list li.item", { hasText: TEST_ITEM });
    await expect(listItem).toBeVisible();
    await listItem.locator('button[hx-post="/list/done"]').click();

    await expect(page.locator("#list").getByText(TEST_ITEM)).toHaveCount(0);
    await expect(page.locator("#cart").getByText(TEST_ITEM)).toBeVisible();
  });

  await test.step("can check out the shopping cart and see all done items be cleared", async () => {
    await page.locator('#cart button[hx-post="/list/finish"]').click();

    await expect(page.locator("#cart").getByText(TEST_ITEM)).toHaveCount(0);
    await expect(page.locator("#list").getByText(TEST_ITEM)).toHaveCount(0);
  });
});
