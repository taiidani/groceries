package models

import (
	"context"
	"testing"
)

func TestRecipe_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		recipe  Recipe
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid recipe with name only",
			recipe: Recipe{
				Name: "Pasta Carbonara",
			},
			wantErr: false,
		},
		{
			name: "valid recipe with name and description",
			recipe: Recipe{
				Name:        "Chicken Curry",
				Description: "A delicious Indian curry",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			recipe: Recipe{
				Name:        "",
				Description: "A recipe with no name",
			},
			wantErr: true,
			errMsg:  "recipe name cannot be empty",
		},
		{
			name: "whitespace only name",
			recipe: Recipe{
				Name:        "   ",
				Description: "Test",
			},
			wantErr: false, // Validation doesn't trim, so this is technically valid
		},
		{
			name: "long name is valid",
			recipe: Recipe{
				Name: "This is a very long recipe name that should still be valid because we don't have a max length",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			err := tt.recipe.Validate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Recipe.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Recipe.Validate() error message = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestRecipeItem_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		recipeItem RecipeItem
		wantID     int
		wantName   string
	}{
		{
			name: "recipe item with all fields",
			recipeItem: RecipeItem{
				ID:       1,
				RecipeID: 100,
				ItemID:   200,
				Quantity: "2 cups",
				ItemName: "Flour",
			},
			wantID:   1,
			wantName: "Flour",
		},
		{
			name: "recipe item without quantity",
			recipeItem: RecipeItem{
				ID:       2,
				RecipeID: 101,
				ItemID:   201,
				ItemName: "Sugar",
			},
			wantID:   2,
			wantName: "Sugar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.recipeItem.ID != tt.wantID {
				t.Errorf("RecipeItem.ID = %v, want %v", tt.recipeItem.ID, tt.wantID)
			}
			if tt.recipeItem.ItemName != tt.wantName {
				t.Errorf("RecipeItem.ItemName = %v, want %v", tt.recipeItem.ItemName, tt.wantName)
			}
		})
	}
}
