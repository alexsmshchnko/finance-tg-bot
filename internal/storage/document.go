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
	Direction   int       `db:"direction"`
}

func (s *PGStorage) GetCategories(username string) (cat []string, err error) {
	err = s.db.Select(&cat, "select trans_cat from public.document where client_id = $1 group by trans_cat order by count(*) desc", username)
	return
}

func (s *PGStorage) postDocument(doc *DBDocument) (err error) {
	tx := s.db.MustBegin()

	sql := "INSERT INTO public.document (trans_date, trans_cat, trans_amount, comment, tg_msg_id, client_id, direction)" +
		"VALUES($1, $2, $3, $4, $5, $6, $7);"
	tx.MustExec(sql, doc.Time, doc.Category, doc.Amount, doc.Description, doc.MsgID, doc.ClientID, doc.Direction)

	return tx.Commit()
}

func (s *PGStorage) PostDoc(time time.Time, category string, amount int, description string, msg_id string, client string) (err error) {
	doc := &DBDocument{
		Time:        time,
		Category:    category,
		Amount:      amount,
		Description: description,
		MsgID:       msg_id,
		ClientID:    client,
	}

	return s.postDocument(doc)
}

func (s *PGStorage) DeleteDoc(msg_id string, client string) (err error) {
	tx := s.db.MustBegin()

	tx.MustExec("DELETE FROM public.document WHERE tg_msg_id = $1 and client_id = $2;", msg_id, client)

	return tx.Commit()
}

func (s *PGStorage) ClearUserHistory(username string) (err error) {
	tx := s.db.MustBegin()

	tx.MustExec("DELETE FROM public.document WHERE client_id = $1;", username)

	return tx.Commit()
}

func (s *PGStorage) LoadDocs(time time.Time, category string, amount int, description string, direction int, client string) (err error) {
	doc := &DBDocument{
		Time:        time,
		Category:    category,
		Amount:      amount,
		Description: description,
		ClientID:    client,
		Direction:   direction,
	}

	return s.postDocument(doc)
}
