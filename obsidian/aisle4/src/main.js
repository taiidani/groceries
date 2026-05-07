import { Notice, Plugin } from "obsidian";
import { DEFAULT_SETTINGS, Aisle4SettingTab } from "./settings.js";
import { findRecipeLeaf, scrapeIngredients } from "./scraper.js";
import { Aisle4Modal } from "./modal.js";
import { verifyItems, addToGroceryList } from "./api.js";

// ─────────────────────────────────────────────────────────────────────────────
// Plugin
// ─────────────────────────────────────────────────────────────────────────────

export default class Aisle4Plugin extends Plugin {
  async onload() {
    await this.loadSettings();

    this.addRibbonIcon("shopping-cart", "Add to grocery list", async () => {
      await this.onRibbonClick();
    });

    this.addSettingTab(new Aisle4SettingTab(this.app, this));
  }

  onunload() {}

  async onRibbonClick() {
    const recipeLeaf = findRecipeLeaf(this.app);
    if (!recipeLeaf) {
      new Notice("Aisle4: Open this note in Recipe View first.", 4000);
      return;
    }

    const ingredients = scrapeIngredients(recipeLeaf.view.contentEl);

    if (ingredients.length === 0) {
      new Notice("Aisle4: No ingredients found in this recipe.");
      return;
    }

    new Aisle4Modal(
      this.app,
      ingredients,
      async (items) => verifyItems(items, this.settings),
      async (selected) => {
        const { added, appended, alreadyOnList, errors } =
          await addToGroceryList(selected, this.settings);

        const parts = [];
        if (added > 0)
          parts.push(`${added} item${added === 1 ? "" : "s"} added`);
        if (appended > 0)
          parts.push(`${appended} item${appended === 1 ? "" : "s"} updated`);
        if (alreadyOnList > 0) parts.push(`${alreadyOnList} already on list`);
        if (errors.length > 0) parts.push(`${errors.length} failed`);

        new Notice(
          `Aisle4: ${parts.join(", ")}.`,
          errors.length > 0 ? 8000 : 4000,
        );

        if (errors.length > 0) {
          console.error("Aisle4 — errors adding items:", errors);
        }
      },
    ).open();
  }

  async loadSettings() {
    this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
  }

  async saveSettings() {
    await this.saveData(this.settings);
  }
}
