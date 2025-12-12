package models

import (
	"context"
	"errors"
	"fmt"
)

type Group struct {
	ID   int
	Name string
}

func (c *Group) Validate(ctx context.Context) error {
	var vErr error

	if len(c.Name) < 3 {
		vErr = errors.Join(vErr, errors.New("provided name needs to be at least 3 characters"))
	}

	return vErr
}

func LoadGroups(ctx context.Context) ([]Group, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, name
FROM "group"
ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Group{}
	for rows.Next() {
		// Load the group
		var data Group
		if err := rows.Scan(&data.ID, &data.Name); err != nil {
			return nil, err
		}

		ret = append(ret, data)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func GetGroup(ctx context.Context, id int) (Group, error) {
	row := db.QueryRowContext(ctx, `
SELECT id, name
FROM "group"
WHERE id = $1`, id)
	if row.Err() != nil {
		return Group{}, row.Err()
	}

	// Load the group
	var data Group
	err := row.Scan(&data.ID, &data.Name)
	return data, err
}

func AddGroup(ctx context.Context, data Group) error {
	_, err := db.ExecContext(ctx, `INSERT INTO "group" (name) VALUES ($1)`, data.Name)
	return err
}

func EditGroup(ctx context.Context, data Group) error {
	_, err := db.ExecContext(ctx, `
UPDATE "group" SET
	name = $2,
WHERE id = $1`, data.ID, data.Name)
	return err
}

func DeleteGroup(ctx context.Context, id int) error {
	// Prevent deletion if group is still in use
	data, err := UsersForGroup(ctx, id)
	if err != nil {
		return fmt.Errorf("could not enumerate group users: %w", err)
	}
	if len(data) > 0 {
		return errors.New("group is still in use")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "DELETE FROM group WHERE id = $1", id)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}
