package storage

import (
	"encoding/json"
	"time"
)

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

type DocExport struct {
	Time        time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int       `json:"trans_amount"`
	Description string    `json:"comment"`
	Direction   string    `json:"direction"`
}

func (s *PGStorage) GetCategories(username string) (cat []string, err error) {
	err = s.db.Select(&cat, "select trans_cat from public.document"+
		" where client_id = $1 group by trans_cat order by count(*) desc", username)
	return
}

func (s *PGStorage) GetSubCategories(username, trans_cat string) (cat []string, err error) {
	err = s.db.Select(&cat, "select lower(comment) from public.document"+
		" where client_id = $1 and trans_cat = $2 and trans_date > current_date - 90"+
		" group by lower(comment) order by count(*) desc limit 10", username, trans_cat)
	return
}

func (s *PGStorage) postDocument(doc *DBDocument) (err error) {
	if doc.Direction == 0 {
		err = s.db.Get(&doc.Direction, "select direction from public.trans_category"+
			" where client_id = $1 and trans_cat = $2 and active = true", doc.ClientID, doc.Category)
		if err != nil {
			return err
		}
	}

	tx := s.db.MustBegin()
	sql := "INSERT INTO public.document (trans_date, trans_cat, trans_amount, comment, tg_msg_id, client_id, direction)" +
		" VALUES($1, $2, $3, $4, $5, $6, $7);"
	tx.MustExec(sql, doc.Time, doc.Category, doc.Amount, doc.Description, doc.MsgID, doc.ClientID, doc.Direction)

	return tx.Commit()
}

func (s *PGStorage) PostDoc(time time.Time, category string, amount int, description string, msg_id string, direction int, client string) (err error) {
	doc := &DBDocument{
		Time:        time,
		Category:    category,
		Amount:      amount,
		Description: description,
		MsgID:       msg_id,
		ClientID:    client,
		Direction:   direction,
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
	tx.MustExec("UPDATE public.trans_category SET active = false WHERE client_id = $1;", username)

	return tx.Commit()
}

func (s *PGStorage) ImportDocs(data []byte, client string) (err error) {
	var docs []DBDocument

	err = json.Unmarshal(data, &docs)
	if err != nil {
		return err
	}

	for _, v := range docs {
		s.postDocument(&DBDocument{
			Time:        v.Time,
			Category:    v.Category,
			Amount:      v.Amount,
			Description: v.Description,
			ClientID:    client,
			Direction:   v.Direction,
		})
		if err != nil {
			return
		}
	}

	tx := s.db.MustBegin()
	sql := "INSERT INTO trans_category(trans_cat, direction, client_id)" +
		" SELECT distinct trans_cat, direction, client_id FROM document WHERE client_id = $1"
	tx.MustExec(sql, client)

	return tx.Commit()
}

// func (s *PGStorage) LoadDocs(time time.Time, category string, amount int, description string, direction int, client string) (err error) {
// 	doc := &DBDocument{
// 		Time:        time,
// 		Category:    category,
// 		Amount:      amount,
// 		Description: description,
// 		ClientID:    client,
// 		Direction:   direction,
// 	}

// 	return s.postDocument(doc)
// }

func (s *PGStorage) Export(client string) (rslt []byte, err error) {
	data, err := s.db.Query("SELECT trans_date, trans_cat, trans_amount, comment, case direction when -1 then 'debit' when 1 then 'credit' else 'other' end as direction"+
		" FROM base.public.document WHERE client_id = $1 ORDER BY 1 DESC", client)
	if err != nil {
		return rslt, err
	}

	expDoc := DocExport{}
	var expDocs []DocExport

	for data.Next() {
		err = data.Scan(&expDoc.Time, &expDoc.Category, &expDoc.Amount, &expDoc.Description, &expDoc.Direction)
		if err != nil {
			return rslt, err
		}
		expDocs = append(expDocs, expDoc)
	}
	return json.Marshal(expDocs)

}
