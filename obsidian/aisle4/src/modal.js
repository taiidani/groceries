import { Modal, Notice } from "obsidian";

// ─────────────────────────────────────────────────────────────────────────────
// Review modal
// ─────────────────────────────────────────────────────────────────────────────

// Status values for each ingredient row after verification:
//   'known'   — name matched in catalog, not yet on the shopping list
//   'on-list' — name matched in catalog and already on the shopping list
//   'new'     — name not found; will be created as a new item on submit
//   null      — not yet verified

export class Aisle4Modal extends Modal {
  constructor(app, ingredients, onVerify, onSubmit) {
    super(app);
    this.onVerify = onVerify;
    this.onSubmit = onSubmit;
    this.mode = "verify"; // 'verify' | 'submit'
    // Mutable working copy — keeps the original scrape result untouched.
    this.items = ingredients.map((ing) => ({
      group: ing.group,
      name: ing.name,
      quantity: ing.quantity,
      original: ing.original,
      // Pre-uncheck ingredients the user has already crossed off in recipe-view.
      included: !ing.doneInRecipeView,
      status: null, // set by handleVerify()
      listItemId: null, // set by handleVerify() for 'on-list' items
      existingQuantity: null, // set by handleVerify() for 'on-list' items
      statusEl: null, // DOM reference, set in onOpen()
      nameInputEl: null, // DOM reference, set in onOpen()
    }));
  }

  onOpen() {
    this.modalEl.addClass("aisle4-modal");
    this.titleEl.setText("Add to Grocery List");

    const { contentEl } = this;

    // ── Ingredient list ───────────────────────────────────────────────────

    const listEl = contentEl.createDiv({ cls: "aisle4-ingredient-list" });
    let renderedGroup = undefined;

    for (const item of this.items) {
      // Emit a group heading whenever the group label changes.
      if (item.group !== renderedGroup) {
        renderedGroup = item.group;
        if (renderedGroup) {
          listEl.createDiv({
            cls: "aisle4-group-heading",
            text: renderedGroup,
          });
        }
      }

      const itemEl = listEl.createDiv({ cls: "aisle4-ingredient-item" });
      const row = itemEl.createDiv({ cls: "aisle4-ingredient-row" });

      // Checkbox — unchecking dims the item and excludes it from submission.
      // Pre-unchecked when the ingredient was already crossed off in recipe-view.
      const checkbox = row.createEl("input");
      checkbox.type = "checkbox";
      checkbox.checked = item.included;
      checkbox.addEventListener("change", () => {
        item.included = checkbox.checked;
        itemEl.toggleClass("aisle4-row-excluded", !checkbox.checked);
      });
      if (!item.included) itemEl.addClass("aisle4-row-excluded");

      // Quantity input (narrow). Quantity changes don't affect catalog
      // matching, so they don't reset the verify state.
      const qtyInput = row.createEl("input", { cls: "aisle4-qty-input" });
      qtyInput.type = "text";
      qtyInput.value = item.quantity;
      qtyInput.placeholder = "—";
      qtyInput.addEventListener("input", () => {
        item.quantity = qtyInput.value;
      });

      // Name input (fills remaining space). Name changes invalidate the
      // verification results and revert the modal to verify mode.
      const nameInput = row.createEl("input", { cls: "aisle4-name-input" });
      nameInput.type = "text";
      nameInput.value = item.name;
      nameInput.addEventListener("input", () => {
        item.name = nameInput.value;
        if (this.mode === "submit") {
          this.setMode("verify");
        }
      });
      item.nameInputEl = nameInput;

      // Status badge — hidden until verification runs.
      item.statusEl = row.createSpan({ cls: "aisle4-item-status" });

      // Original line as displayed in recipe-view — reference for correcting parsed fields.
      if (item.original) {
        itemEl.createDiv({
          cls: "aisle4-ingredient-original",
          text: item.original,
        });
      }
    }

    // ── Footer ────────────────────────────────────────────────────────────

    const footer = contentEl.createDiv({ cls: "aisle4-modal-footer" });

    footer
      .createEl("button", { text: "Cancel" })
      .addEventListener("click", () => this.close());

    // Single action button that alternates between "Verify" and "Add to List".
    this.actionBtn = footer.createEl("button", {
      text: "Verify",
      cls: "mod-cta",
    });
    this.actionBtn.addEventListener("click", async () => {
      if (this.mode === "verify") {
        await this.handleVerify();
      } else {
        await this.handleSubmit();
      }
    });
  }

  // ── Mode management ───────────────────────────────────────────────────────

  setMode(mode) {
    this.mode = mode;
    if (mode === "verify") {
      // Clear all status badges and verification data so the user knows they
      // need to re-verify before submitting.
      for (const item of this.items) {
        item.status = null;
        item.listItemId = null;
        item.existingQuantity = null;
        this.updateStatusEl(item);
      }
      this.actionBtn.disabled = false;
      this.actionBtn.textContent = "Verify";
    } else {
      // 'submit'
      this.actionBtn.disabled = false;
      this.actionBtn.textContent = "Add to List";
    }
  }

  // Updates the status badge element for a single item.
  updateStatusEl(item) {
    if (!item.statusEl) return;
    const el = item.statusEl;
    // Reset to base class only.
    el.className = "aisle4-item-status";
    el.textContent = "";
    if (!item.status) return;

    const statusConfig = {
      known: { cls: "aisle4-item-status--known", text: "Known" },
      "on-list": { cls: "aisle4-item-status--on-list", text: "On list" },
      new: { cls: "aisle4-item-status--new", text: "New" },
    };
    const config = statusConfig[item.status];
    if (config) {
      el.classList.add(config.cls);
      el.textContent = config.text;
    }
  }

  // ── Action handlers ───────────────────────────────────────────────────────

  async handleVerify() {
    this.actionBtn.disabled = true;
    this.actionBtn.textContent = "Verifying…";
    try {
      const results = await this.onVerify(this.items);
      results.forEach(
        ({ status, canonicalName, listItemId, existingQuantity }, i) => {
          const item = this.items[i];
          item.status = status;
          item.listItemId = listItemId;
          item.existingQuantity = existingQuantity;
          // Silently correct casing to match the API's stored name.
          if (canonicalName !== null && canonicalName !== item.name) {
            item.name = canonicalName;
            if (item.nameInputEl) item.nameInputEl.value = canonicalName;
          }
          this.updateStatusEl(item);
        },
      );
      this.setMode("submit");
    } catch (err) {
      new Notice("Aisle4: Verification failed — " + err.message, 6000);
      this.actionBtn.disabled = false;
      this.actionBtn.textContent = "Verify";
    }
  }

  async handleSubmit() {
    const selected = this.items.filter((it) => it.included && it.name.trim());
    if (selected.length === 0) {
      new Notice("Aisle4: No items selected.");
      return;
    }
    this.actionBtn.disabled = true;
    this.actionBtn.textContent = "Adding…";
    try {
      await this.onSubmit(selected);
    } finally {
      this.close();
    }
  }

  onClose() {
    this.contentEl.empty();
  }
}
