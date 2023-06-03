package db

import (
	"context"
	"fmt"
	"time"
)

// operations with files
type CreateFileParams struct {
	Owner       int64     `json:"owner"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	LocationWav string    `json:"location_wav"`
	LocationTxt string    `json:"location_txt"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
	Guid        string    `json:"guid"`
}

const createFile = `-- name: CreateFile :one
INSERT INTO files (
	owner,
	name,
	location,
	location_wav,
	location_txt,
	created_at,
	status,
	guid
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id, owner, name, location, location_wav, location_txt, created_at, status, guid
`

func (q *Queries) CreateFile(ctx context.Context, arg CreateFileParams) (File, error) {
	row := q.db.QueryRowContext(ctx, createFile, arg.Owner, arg.Name, arg.Location, arg.LocationWav, arg.LocationTxt, arg.CreatedAt, arg.Status, arg.Guid)
	var i File
	err := row.Scan(
		&i.Id,
		&i.Owner,
		&i.Name,
		&i.Location,
		&i.LocationWav,
		&i.LocationTxt,
		&i.CreatedAt,
		&i.Status,
		&i.Guid,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i, err
}

const getFile = `-- name: GetFile :one
SELECT id, owner, name, location, location_wav, location_txt, created_at, status, guid FROM files
WHERE id = $1 AND owner = $2 LIMIT 1
`

func (q *Queries) GetFile(ctx context.Context, id int64, owner int64) (File, error) {
	row := q.db.QueryRowContext(ctx, getFile, id, owner)
	var i File
	err := row.Scan(
		&i.Id,
		&i.Owner,
		&i.Name,
		&i.Location,
		&i.LocationWav,
		&i.LocationTxt,
		&i.CreatedAt,
		&i.Status,
		&i.Guid,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i, err
}

const listFiles = `--name: ListFiles :many
SELECT id, owner, name, location, location_wav, location_txt, created_at, status, guid FROM files
WHERE owner = $3
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListFilesParams struct {
	Limit  int32 `json:"id"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListFiles(ctx context.Context, arg ListFilesParams, owner int64) ([]File, error) {
	rows, err := q.db.QueryContext(ctx, listFiles, arg.Limit, arg.Offset, owner)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	var items []File
	for rows.Next() {
		var i File
		if err := rows.Scan(
			&i.Id,
			&i.Owner,
			&i.Name,
			&i.Location,
			&i.LocationWav,
			&i.LocationTxt,
			&i.CreatedAt,
			&i.Status,
			&i.Guid,
		); err != nil {
			fmt.Print(err)
			return nil, err
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

const updateFile = `--name: UpdateFile :one
UPDATE files
SET name = $2, status = $3
WHERE id = $1
RETURNING id, name, location, location_wav, location_txt, status
`

type UpdateFileParams struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (q *Queries) UpdateFile(ctx context.Context, arg UpdateFileParams) (File, error) {
	row := q.db.QueryRowContext(ctx, updateFile, arg.Id, arg.Name, arg.Status)
	var i File
	err := row.Scan(
		&i.Id,
		&i.Name,
		&i.Location,
		&i.LocationWav,
		&i.LocationTxt,
		&i.Status,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i, err
}

const deleteFile = `--name: DeleteFile :exec
DELETE FROM files
WHERE id = $1
`

func (q *Queries) DeleteFile(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteFile, id)
	if err != nil {
		fmt.Print(err)
		return err
	}
	return nil
}

const getLastFileId = `--name: GetLastAccountId :one
SELECT MAX(id) AS max_id
FROM files;
`

func (q *Queries) GetLastFileID(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, getLastFileId)

	var i File
	err := row.Scan(
		&i.Id,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i.Id, err
}

const getFileByGuid = `--name: GetFileByGuid :one 
SELECT id, owner, name, location, location_wav, location_txt, created_at, status, guid FROM files
WHERE guid = $1 LIMIT 1
`

func (q *Queries) GetFileByGuid(ctx context.Context, guid string) (File, error) {
	row := q.db.QueryRowContext(ctx, getFileByGuid, guid)
	var i File
	err := row.Scan(
		&i.Id,
		&i.Owner,
		&i.Name,
		&i.Location,
		&i.LocationWav,
		&i.LocationTxt,
		&i.CreatedAt,
		&i.Status,
		&i.Guid,
	)
	if err != nil {
		fmt.Print(err)
	}
	return i, err
}
