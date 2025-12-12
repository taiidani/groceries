package models

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"io"
)

type User struct {
	ID    int
	Name  string
	Admin bool
}

func (u *User) Validate(ctx context.Context) error {
	var vErr error
	if u.Name == "" {
		vErr = errors.Join(vErr, errors.New("username must be valid"))
	}

	return vErr
}

func LoadUsers(ctx context.Context) ([]User, error) {
	rows, err := db.QueryContext(ctx, `
SELECT u.id, u.name, u.admin
FROM "user" u
ORDER BY u.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []User{}
	for rows.Next() {
		// Load the user
		user := User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.Admin); err != nil {
			return nil, err
		}

		ret = append(ret, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func GetUser(ctx context.Context, id int) (User, error) {
	ret := User{}
	err := db.QueryRowContext(ctx, `
SELECT u.id, u.name, u.admin
FROM "user" u
WHERE u.id = $1`, id).
		Scan(&ret.ID, &ret.Name, &ret.Admin)
	if err != nil {
		return ret, err
	}

	return ret, err
}

func GetUserByCredentials(ctx context.Context, name string, password string) (User, error) {
	// Super secret, just between us
	const expected = "ab77936ff6728921c550adb7fc338623"

	hasher := md5.New()
	io.WriteString(hasher, password)
	sum := fmt.Sprintf("%x", hasher.Sum(nil))
	if sum != expected {
		return User{}, errors.New("invalid password")
	}

	ret := User{}
	err := db.QueryRowContext(ctx, `
SELECT u.id, u.name, u.admin
FROM "user" u
WHERE u.name = $1`, name).
		Scan(&ret.ID, &ret.Name, &ret.Admin)

	if errors.Is(err, sql.ErrNoRows) {
		return ret, errors.New("user not found")
	}

	return ret, err
}

func AddUser(ctx context.Context, u User) error {
	if err := u.Validate(ctx); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if u.ID != 0 {
		return errors.New("resource already had an id assigned")
	}

	u.ID, err = insertWithID(ctx, tx,
		`INSERT INTO "user" (name, admin) VALUES ($1, $2) RETURNING id`,
		u.Name,
		u.Admin,
	)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}

func EditUser(ctx context.Context, u User) error {
	if err := u.Validate(ctx); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
UPDATE "user" SET
	name = $2,
	admin = $3
WHERE id = $1`, u.ID, u.Name, u.Admin)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}

func DeleteUser(ctx context.Context, id int) error {
	_, err := db.ExecContext(ctx, `DELETE FROM "user" WHERE id = $1`, id)
	return err
}
