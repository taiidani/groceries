package models

import (
	"context"
	"testing"
)

func TestListItem_Validate(t *testing.T) {
	tests := []struct {
		name     string
		listItem ListItem
		wantErr  bool
	}{
		{
			name: "valid list item",
			listItem: ListItem{
				ID:         1,
				ItemID:     1,
				CategoryID: "produce",
				Quantity:   "2",
				Done:       false,
				Name:       "Apples",
			},
			wantErr: false,
		},
		{
			name: "valid list item - done",
			listItem: ListItem{
				ID:         2,
				ItemID:     2,
				CategoryID: "dairy",
				Quantity:   "1 gallon",
				Done:       true,
				Name:       "Milk",
			},
			wantErr: false,
		},
		{
			name: "empty list item",
			listItem: ListItem{
				ID:         0,
				ItemID:     0,
				CategoryID: "",
				Quantity:   "",
				Done:       false,
				Name:       "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := tt.listItem.Validate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListItem.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
