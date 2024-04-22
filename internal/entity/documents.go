package entity

import (
	"database/sql"
	"time"
)

type UserStats struct {
	ClientID      int `json:"client_id"`
	AvgIncome     int `json:"income"`
	MonthWrkHours int `json:"month_work_hours"`
	AvgExpenses   int `json:"avg_expenses"`
	LowExpenses   int `json:"low_expenses"`
}

type DocumentExport struct {
	Time        time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int       `json:"trans_amount"`
	ClientID    string    `json:"client_id"`
	MsgID       string    `json:"tg_msg_id"`
	Description string    `json:"comment"`
	Direction   string    `json:"direction"`
}

type TransCatLimit struct {
	Category  sql.NullString `json:"trans_cat"`
	Direction sql.NullInt16  `json:"direction"`
	ClientID  sql.NullString `json:"client_id"`
	Active    sql.NullBool   `json:"active"`
	Limit     sql.NullInt64  `json:"trans_limit"`
	Balance   sql.NullInt64  `json:"trans_balance"`
}

type Document struct {
	RecTime     time.Time `db:"rec_time"     json:"rec_time"`
	TransDate   time.Time `db:"trans_date"   json:"trans_date"`
	Category    string    `db:"trans_cat"    json:"trans_cat"`
	Amount      int64     `db:"trans_amount" json:"trans_amount"`
	Description string    `db:"comment"      json:"comment"`
	MsgID       string    `db:"tg_msg_id"`
	ChatID      string    `db:"tg_chat_id"`
	ClientID    string    `db:"client_id"`
	Direction   int16     `db:"direction"    json:"direction"`
}
