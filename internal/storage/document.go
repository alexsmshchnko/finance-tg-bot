package storage

import (
	"time"
)

type DBDocument struct {
	ID          int64     `db:"id"`
	Time        time.Time `db:"trans_date"`
	Category    string    `db:"trans_cat"`
	Amount      int       `db:"trans_amount"`
	Description string    `db:"comment"`
	MsgID       string    `db:"tg_msg_id"`
	ClientID    string    `db:"client_id"`
}

func (s *PGStorage) GetCategories(username string) (cat []string, err error) {
	err = s.db.Select(&cat, "select trans_cat from public.document where client_id = $1 group by trans_cat order by count(*) desc", username)
	return
}

func (s *PGStorage) PostDocument(doc *DBDocument) (err error) {
	tx := s.db.MustBegin()

	sql := "INSERT INTO public.document (trans_date, trans_cat, trans_amount, comment, tg_msg_id, client_id)" +
		"VALUES($1, $2, $3, $4, $5, $6);"
	tx.MustExec(sql, doc.Time, doc.Category, doc.Amount, doc.Description, doc.MsgID, doc.ClientID)

	return tx.Commit()
}
