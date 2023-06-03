package db

import (
	"context"
	"fmt"
	"time"
)

type CreateAccountParams struct {
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	Session   string    `json:"session"`
}

const createAccount = `-- name: CreateAccount :one
INSERT INTO users (
	login,
	password,
	address,
	created_at,
	session
) VALUES (
	$1, $2, $3, $4, $5
) RETURNING id, login, password, address, created_at, session
`

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount, arg.Login, arg.Password, arg.Address, time.Now(), arg.Session)
	var i Account
	err := row.Scan(
		&i.Id,
		&i.Login,
		&i.Password,
		&i.Address,
		&i.CreatedAt,
		&i.Session,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i, err
}

const getAccount = `--name: GetAccount :one
SELECT id, login, password, address, created_at, session FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetAccount(ctx context.Context, id int64) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccount, id)
	var i Account
	err := row.Scan(
		&i.Id,
		&i.Login,
		&i.Password,
		&i.Address,
		&i.CreatedAt,
		&i.Session,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i, err
}

const getAccountByLogin = `--name: GetAccount :one
SELECT id, login, password, address, created_at, session FROM users
WHERE login = $1 LIMIT 1
`

func (q *Queries) GetAccountByLogin(ctx context.Context, Login interface{}) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccountByLogin, Login)
	var i Account
	err := row.Scan(
		&i.Id,
		&i.Login,
		&i.Password,
		&i.Address,
		&i.CreatedAt,
		&i.Session,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i, err
}

const listAccounts = `--name: ListAccounts :many
SELECT id, login, password, address, created_at, session FROM users
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListAccountsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListAccounts(ctx context.Context, arg ListAccountsParams) ([]Account, error) {
	rows, err := q.db.QueryContext(ctx, listAccounts, arg.Limit, arg.Offset)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	defer rows.Close()

	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.Id,
			&i.Login,
			&i.Password,
			&i.Address,
			&i.CreatedAt,
			&i.Session,
		); err != nil {
			fmt.Print(err)
		}
		items = append(items, i)
	}

	if err := rows.Close(); err != nil {
		fmt.Print(err)
		return nil, err
	}
	if err := rows.Err(); err != nil {
		fmt.Print(err)
		return nil, err
	}
	return items, nil
}

const updateAccount = `--name: UpdateAccount :one
UPDATE users
SET login = $2, password = $3, address = $4, session = $5
WHERE id = $1
RETURNING id, login, password, address, session
`

type UpdateAccountParams struct {
	Id       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Address  string `json:"address"`
	Session  string `json:"session"`
}

func (q *Queries) UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, updateAccount, arg.Id, arg.Login, arg.Password, arg.Address, arg.Session)

	var i Account
	err := row.Scan(
		&i.Id,
		&i.Login,
		&i.Password,
		&i.Address,
		&i.Session,
	)

	if err != nil {
		fmt.Print(err)
	}

	return i, err
}

const deleteAccount = `--name: DeleteAccount :exec
DELETE FROM users
WHERE id = $1
`

func (q *Queries) DeleteAccount(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteAccount, id)
	if err != nil {
		fmt.Print(err)
	}
	return err
}

const getLast = `--name: GetLastAccountId :one
SELECT MAX(id) AS max_id
FROM users;
`

func (q *Queries) GetLastAccountID(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, getLast)

	var i Account
	err := row.Scan(
		&i.Id,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i.Id, err
}

const getAccountBySession = `--name GetUserBySession :one
SELECT id, login, password, address, created_at, session FROM users
WHERE session = $1 LIMIT 1
`

func (q *Queries) GetAccountBySession(ctx context.Context, session interface{}) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccountBySession, session)
	var i Account
	err := row.Scan(
		&i.Id,
		&i.Login,
		&i.Password,
		&i.Address,
		&i.CreatedAt,
		&i.Session,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i, err

}
