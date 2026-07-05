// @ts-check
const { defineConfig, devices } = require("@playwright/test");

module.exports = defineConfig({
  testDir: "./e2e",

  // The tests share a single instance of the app and mutate the same
  // database state, so they must not run concurrently against each other.
  fullyParallel: false,
  workers: 1,

  forbidOnly: !!process.env.CI,
  // Retries are intentionally disabled: these tests mutate shared database
  // state (the grocery list), so a retried test would not start from a
  // clean slate.
  retries: 0,
  reporter: process.env.CI ? [["github"], ["list"]] : "list",

  use: {
    baseURL: process.env.BASE_URL || "http://localhost:3000",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
