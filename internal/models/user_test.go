package models

import (
	"context"
	"testing"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				ID:    1,
				Name:  "testuser",
				Admin: false,
			},
			wantErr: false,
		},
		{
			name: "valid admin user",
			user: User{
				ID:    2,
				Name:  "admin",
				Admin: true,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			user: User{
				ID:    3,
				Name:  "",
				Admin: false,
			},
			wantErr: true,
		},
		{
			name: "user with only ID",
			user: User{
				ID:    4,
				Name:  "",
				Admin: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := tt.user.Validate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
