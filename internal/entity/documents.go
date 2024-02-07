package entity

import "time"

type Document struct {
	Time        time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int       `json:"trans_amount"`
	Description string    `json:"comment"`
	Direction   string    `json:"direction"`
}
