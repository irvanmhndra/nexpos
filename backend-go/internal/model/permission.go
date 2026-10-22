package model

import "time"

type Permission struct {
	ID          int64     `db:"id"`
	Code        string    `db:"code"`
	Name        string    `db:"name"`
	Module      string    `db:"module"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}
