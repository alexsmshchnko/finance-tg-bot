package entity

import (
	"database/sql"
	"time"
)

type DocumentExport struct {
	Time        time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int       `json:"trans_amount"`
	ClientID    string    `json:"client_id"`
	MsgID       string    `json:"tg_msg_id"`
	Description string    `json:"comment"`
	Direction   string    `json:"direction"`
}

type DBDocument struct {
	ID          *int64     `db:"id"`
	Time        *time.Time `db:"trans_date"   json:"trans_date"`
	Category    *string    `db:"trans_cat"    json:"trans_cat"`
	Amount      *int       `db:"trans_amount" json:"trans_amount"`
	Description *string    `db:"comment"      json:"comment"`
	MsgID       *string    `db:"tg_msg_id"`
	ClientID    *string    `db:"client_id"`
	Direction   *int       `db:"direction"    json:"direction"`
}

type TransCatLimit struct {
	Category  sql.NullString `db:"trans_cat"`
	Direction sql.NullInt16  `db:"direction"`
	ClientID  sql.NullString `db:"client_id"`
	Active    sql.NullBool   `db:"active"`
	Limit     sql.NullInt64  `db:"trans_limit"`
	Balance   sql.NullInt64  `db:"trans_balance"`
}

type Document struct {
	TransDate   time.Time `db:"trans_date"   json:"trans_date"`
	Category    string    `db:"trans_cat"    json:"trans_cat"`
	Amount      int64     `db:"trans_amount" json:"trans_amount"`
	Description string    `db:"comment"      json:"comment"`
	MsgID       string    `db:"tg_msg_id"`
	ChatID      string    `db:"tg_chat_id"`
	ClientID    string    `db:"client_id"`
	Direction   int16     `db:"direction"    json:"direction"`
}
