package model

import "time"

type Customer struct {
	ID        int64     `db:"id" json:"id"`
	CompanyID int64     `db:"company_id" json:"company_id"`
	Code      string    `db:"code" json:"code"`
	Name      string    `db:"name" json:"name"`
	Phone     *string   `db:"phone" json:"phone"`
	Email     *string   `db:"email" json:"email"`
	IsMember  bool      `db:"is_member" json:"is_member"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
