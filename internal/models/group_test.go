package models

import (
	"context"
	"testing"
)

func TestGroup_Validate(t *testing.T) {
	tests := []struct {
		name    string
		group   Group
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid group",
			group: Group{
				ID:   0,
				Name: "Family",
			},
			wantErr: false,
		},
		{
			name: "valid group with longer name",
			group: Group{
				ID:   0,
				Name: "Household Members",
			},
			wantErr: false,
		},
		{
			name: "name too short - 2 chars",
			group: Group{
				ID:   0,
				Name: "AB",
			},
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
		{
			name: "name too short - 1 char",
			group: Group{
				ID:   0,
				Name: "A",
			},
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
		{
			name: "empty name",
			group: Group{
				ID:   0,
				Name: "",
			},
			wantErr: true,
			errMsg:  "at least 3 characters",
		},
		{
			name: "exactly 3 chars - valid",
			group: Group{
				ID:   0,
				Name: "ABC",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := tt.group.Validate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Group.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
