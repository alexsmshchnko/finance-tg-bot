package entity

import "time"

type DocumentExport struct {
	Time        time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int       `json:"trans_amount"`
	Description string    `json:"comment"`
	Direction   string    `json:"direction"`
}

type DBDocument struct {
	ID          int64     `db:"id"`
	Time        time.Time `db:"trans_date"   json:"trans_date"`
	Category    string    `db:"trans_cat"    json:"trans_cat"`
	Amount      int       `db:"trans_amount" json:"trans_amount"`
	Description string    `db:"comment"      json:"comment"`
	MsgID       string    `db:"tg_msg_id"`
	ClientID    string    `db:"client_id"`
	Direction   int       `db:"direction"    json:"direction"`
}