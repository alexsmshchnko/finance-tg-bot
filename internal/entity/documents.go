package entity

import (
	"time"
)

type UserStats struct {
	UserId        int `json:"client_id"`
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
	Category    string `json:"trans_cat"`
	Direction   int8   `json:"direction"`
	UserId      int    `json:"client_id,omitempty"`
	Active      bool   `json:"active,omitempty"`
	Limit       int    `json:"trans_limit"`
	LimitText   string `json:",omitempty"`
	Balance     int    `json:"trans_balance"`
	BalanceText string `json:",omitempty"`
}

type Document struct {
	RecTime     time.Time `json:"rec_time"`
	TransDate   time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int64     `json:"trans_amount"`
	Description string    `json:"comment"`
	MsgID       string    `json:"msg_id"`
	ChatID      string    `json:"chat_id"`
	UserId      int       `json:"client_id"`
	Direction   int8      `json:"direction"`
}
