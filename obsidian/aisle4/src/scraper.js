// Ingredient scraping from lachholden/obsidian-recipe-view's live DOM.
//
// Attribution: relies on the DOM produced by lachholden/obsidian-recipe-view
// (MIT License) https://github.com/lachholden/obsidian-recipe-view
//   - CheckableIngredientList renders ingredient <ul> elements into .column-side
//     (two-column layout) or as <ul class="bullets"> (one-column layout).
//   - Quantities are wrapped in <span data-qty> elements by injectQuantities().
//   - Sub-headings are rendered as <h3 class="recipe-leaf"> elements.

// Returns the recipe-view leaf for the currently active file, or null.
export function findRecipeLeaf(app) {
	const activeFile = app.workspace.getActiveFile();
	if (!activeFile) return null;

	return app.workspace
		.getLeavesOfType('recipe-view')
		.find(leaf => leaf.view?.file?.path === activeFile.path)
		?? null;
}

// Scrapes parsed ingredients from a recipe-view contentEl.
// Returns an array of { group, name, quantity, original, doneInRecipeView }.
export function scrapeIngredients(contentEl) {
	const sideColumn = contentEl.querySelector('.column-side');
	if (sideColumn) {
		// Two-column: all ingredient lists live in .column-side as plain <ul>
		return extractIngredients(sideColumn, 'ul');
	} else {
		// One-column: recipe-view passes bullets=true to CheckableIngredientList,
		// giving ingredient lists a distinctive ul.bullets class absent from
		// instruction/notes lists rendered as plain RecipeLeaf elements.
		const mainColumn = contentEl.querySelector('.column-main') ?? contentEl;
		return extractIngredients(mainColumn, 'ul.bullets');
	}
}

// Walks a column element in document order, tracking <h3 class="recipe-leaf">
// as group labels and extracting ingredient items from matching <ul> elements.
function extractIngredients(columnEl, ulSelector) {
	const ingredients = [];
	let currentGroup = null;

	for (const el of columnEl.querySelectorAll(`h3.recipe-leaf, ${ulSelector}`)) {
		if (el.matches('h3.recipe-leaf')) {
			currentGroup = el.textContent?.trim() ?? null;
			continue;
		}

		// It's a matching <ul> — each <li> is one ingredient.
		for (const li of el.querySelectorAll('li')) {
			// recipe-view moves the rendered li content into a .leaf > div > .recipe-leaf
			// subtree via RecipeLeaf.svelte's onMount. Query from .leaf if present.
			const source = li.querySelector('.leaf') ?? li;

			// The original ingredient line as displayed in recipe-view (scaled quantities,
			// unicode fractions included) — captured before any span removal.
			const original = (source.textContent ?? '').replace(/\s+/g, ' ').trim();

			// recipe-view stores checked state on its own <input type="checkbox">.
			// If the user has already crossed off this ingredient while cooking, default
			// it to unchecked in our modal so it won't be re-added to the list.
			const doneInRecipeView =
				li.querySelector('input[type="checkbox"]')?.checked ?? false;

			// Quantity: join non-parenthetical [data-qty] spans only.
			// Parenthetical spans (e.g. "35 g" in "1/4 cup (35 g)") are alternate-unit
			// annotations — not useful as grocery quantities — and are excluded.
			const quantity = Array.from(source.querySelectorAll('[data-qty]'))
				.filter(s => !isParentheticalQty(s))
				.map(s => s.textContent?.trim())
				.filter(Boolean)
				.join(' ');

			// Name: clone the source, then strip every [data-qty] span.
			// Capture sibling references *before* any removal — removing a span changes
			// its neighbours' sibling pointers, which would break parenthesis detection.
			const clone = source.cloneNode(true);
			const spanInfos = Array.from(clone.querySelectorAll('[data-qty]')).map(s => ({
				span: s,
				prev: s.previousSibling,
				next: s.nextSibling,
				parenthetical: isParentheticalQty(s),
			}));

			for (const { span, prev, next, parenthetical } of spanInfos) {
				if (parenthetical) {
					// Also remove the surrounding "(" and ")" from the adjacent text nodes.
					if (prev?.nodeType === 3)
						prev.textContent = prev.textContent.replace(/\s*\(\s*$/, '');
					if (next?.nodeType === 3)
						next.textContent = next.textContent.replace(/^\s*\)\s*/, '');
				}
				span.remove();
			}

			const name = (clone.textContent ?? '').replace(/\s+/g, ' ').trim();

			if (name || quantity) {
				ingredients.push({ group: currentGroup, name, quantity, original, doneInRecipeView });
			}
		}
	}

	return ingredients;
}

// Returns true when a [data-qty] span is flanked by "(" and ")" text nodes,
// indicating it's an alternate-unit annotation (e.g. "35 g" in "1/4 cup (35 g)")
// rather than a primary quantity. nodeType 3 === Node.TEXT_NODE.
function isParentheticalQty(span) {
	const prev = span.previousSibling;
	const next = span.nextSibling;
	return (
		prev?.nodeType === 3 && /\(\s*$/.test(prev.textContent ?? '') &&
		next?.nodeType === 3 && /^\s*\)/.test(next.textContent ?? '')
	);
}
