package models

import (
	"context"
	"errors"
)

func UsersForGroup(ctx context.Context, groupID int) ([]User, error) {
	rows, err := db.QueryContext(ctx, `SELECT user_id FROM user_group WHERE group_id = $1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []User{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		user, err := GetUser(ctx, id)
		if err != nil {
			return nil, err
		}
		ret = append(ret, user)
	}
	return ret, rows.Err()
}

func AddUserToGroup(ctx context.Context, user User, group Group) error {
	if user.ID == 0 {
		return errors.New("user ID is required")
	}
	if group.ID == 0 {
		return errors.New("group ID is required")
	}

	_, err := db.ExecContext(ctx, `INSERT INTO user_group (user_id, group_id) VALUES ($1, $2)`, user.ID, group.ID)
	return err
}

func RemoveUserFromGroup(ctx context.Context, user User, group Group) error {
	if user.ID == 0 {
		return errors.New("user ID is required")
	}
	if group.ID == 0 {
		return errors.New("group ID is required")
	}

	_, err := db.ExecContext(ctx, `DELETE FROM user_group WHERE user_id = $1 AND group_id = $2`, user.ID, group.ID)
	return err
}
