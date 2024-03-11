package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"finance-tg-bot/internal/entity"
	"finance-tg-bot/pkg/postgres"
	"finance-tg-bot/pkg/repository"
	"finance-tg-bot/pkg/ydb"
)

type Repo struct {
	*postgres.Postgres
	*ydb.Ydb
}

func New(pg *postgres.Postgres, ydb *ydb.Ydb) *Repo {
	return &Repo{Postgres: pg, Ydb: ydb}
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

type TransCat struct {
	ID        sql.NullInt64  `db:"id"`
	Category  sql.NullString `db:"trans_cat"`
	Direction sql.NullInt16  `db:"direction"`
	ClientID  sql.NullString `db:"client_id"`
	Active    sql.NullBool   `db:"active"`
	Limit     sql.NullInt64  `db:"trans_limit"`
}

func (s *Repo) PostDocument(ctx context.Context, doc *entity.Document) (err error) {
	dbdoc := &repository.DBDocument{
		TransDate:   sql.NullTime{Time: time.Time{}, Valid: false},
		Category:    sql.NullString{String: doc.Category, Valid: true},
		Amount:      sql.NullInt64{Int64: doc.Amount, Valid: true},
		Description: sql.NullString{String: doc.Description, Valid: true},
		MsgID:       sql.NullString{String: doc.MsgID, Valid: true},
		ChatID:      sql.NullString{String: doc.ChatID, Valid: true},
		ClientID:    sql.NullString{String: doc.ClientID, Valid: true},
		Direction:   sql.NullInt16{Int16: 0, Valid: false},
	}
	return repository.PostDocument(*s.Ydb, ctx, dbdoc)
}

func (s *Repo) DeleteDocument(ctx context.Context, doc *entity.Document) (err error) {
	dbdoc := &repository.DBDocument{
		MsgID:    sql.NullString{String: doc.MsgID, Valid: true},
		ChatID:   sql.NullString{String: doc.ChatID, Valid: true},
		ClientID: sql.NullString{String: doc.ClientID, Valid: true},
	}
	return repository.DeleteDocument(*s.Ydb, ctx, dbdoc)

}

func (s *Repo) GetCategories(username string) (cat []string, err error) {
	err = s.Select(&cat, `
	select d.trans_cat
      from public.document d
      join public.trans_category tc on (tc.trans_cat = d.trans_cat
                                    and tc.client_id = d.client_id
									and tc.active = true)
     where d.client_id = $1 group by d.trans_cat order by count(*) desc`, username)
	return
}

func (s *Repo) GetCats(ctx context.Context, username, limit string) (cat []entity.TransCatLimit, err error) {
	var sql string
	switch limit {
	case "setting":
		sql = `
		select tc.trans_cat, tc.direction, tc.trans_limit
		  from public.trans_category tc
		  left join public.document d on (tc.trans_cat = d.trans_cat
									  and tc.client_id = d.client_id)
		 where tc.active = true
		   and tc.client_id = $1
		 group by tc.trans_cat, tc.direction, tc.trans_limit
		 order by count(*) desc`
	case "balance":
		sql = `
		select tc.trans_cat, tc.direction
		      ,tc.trans_limit - sum(case when d.trans_date >= date_trunc('month', current_date) then d.trans_amount else 0 end) as trans_limit
	      from public.trans_category tc
	      left join public.document d on (d.trans_cat = tc.trans_cat
								      and d.client_id = tc.client_id
								      and d.trans_date between date_trunc('month', current_date - interval '3' month)
									                       and date_trunc('day', current_date + interval '1' day) - interval '1' second)
         where tc.active = true
	       and tc.client_id = $1
         group by tc.trans_cat, tc.direction, tc.trans_limit
         order by count(d.*) desc`
	}
	err = s.Select(&cat, sql, username)
	return
}

func (s *Repo) GetSubCategories(username, trans_cat string) (cat []string, err error) {
	var res []sql.NullString
	err = s.Select(&res, "select lower(comment) from public.document"+
		" where client_id = $1 and trans_cat = $2 and trans_date > current_date - 90"+
		" group by lower(comment) order by count(*) desc limit 10", username, trans_cat)

	cat = make([]string, 0, len(res))
	for _, v := range res {
		cat = append(cat, v.String)
	}
	return
}

func (s *Repo) postDocument(doc *DBDocument) (err error) {
	if *doc.Direction == 0 {
		err = s.Get(&doc.Direction, "select direction from public.trans_category"+
			" where client_id = $1 and trans_cat = $2 and active = true", doc.ClientID, doc.Category)
		if err != nil {
			return err
		}
	}

	tx := s.MustBegin()
	sql := "INSERT INTO public.document (trans_date, trans_cat, trans_amount, comment, tg_msg_id, client_id, direction)" +
		" VALUES($1, $2, $3, $4, $5, $6, $7);"
	tx.MustExec(sql, doc.Time, doc.Category, doc.Amount, doc.Description, doc.MsgID, doc.ClientID, doc.Direction)

	return tx.Commit()
}

func (s *Repo) getTransCat(category string, active bool, client string) (*TransCat, error) {
	var transCat TransCat
	err := s.Get(&transCat, "select * from trans_category tc where trans_cat = $1 and client_id = $2", category, client)
	return &transCat, err
}

func (s *Repo) createTransCat(tc *TransCat) (err error) {
	tx := s.MustBegin()
	sql := `INSERT INTO public.trans_category(trans_cat, direction, client_id, active)
		    VALUES(lower($1), $2, $3, true)`
	tx.MustExec(sql, tc.Category, tc.Direction, tc.ClientID)

	sql = `INSERT INTO public.document(trans_cat, trans_amount, client_id, direction)
	       select tc.trans_cat, 0, tc.client_id, tc.direction
	         from trans_category tc
            where trans_cat = lower($1) and direction = $2
	          and client_id = $3 and active = true
			limit 1`
	tx.MustExec(sql, tc.Category, tc.Direction, tc.ClientID)

	return tx.Commit()
}

func (s *Repo) updateTransCatLimit(tc *TransCat) (err error) {
	tx := s.MustBegin()
	sql := `UPDATE public.trans_category SET trans_limit = $1
	         WHERE trans_cat = $2
			   AND client_id = $3
			   AND active = true`
	tx.MustExec(sql, tc.Limit.Int64, tc.Category.String, tc.ClientID.String)

	return tx.Commit()
}

func (s *Repo) disableTransCat(tc *TransCat) (err error) {
	tx := s.MustBegin()
	sql := `UPDATE public.trans_category SET active = false
	         WHERE trans_cat = $1
			   AND client_id = $2
			   AND active = true`
	tx.MustExec(sql, tc.Category.String, tc.ClientID.String)

	return tx.Commit()
}

func (s *Repo) EditDocCategory(ctx context.Context, tc entity.TransCatLimit) (err error) {
	dbTCL := &repository.TransCat{
		Category:  tc.Category,
		Direction: tc.Direction,
		ClientID:  tc.ClientID,
		Active:    tc.Active,
	}
	if tc.Limit.Valid {
		dbTCL.Limit = tc.Limit
	}

	return repository.EditCategory(*s.Ydb, ctx, dbTCL)
}

func (s *Repo) EditCategory(tc entity.TransCatLimit, client string) (err error) {
	_, err = s.getTransCat(tc.Category.String, tc.Active.Bool, client)
	t := &TransCat{
		Category:  tc.Category,
		Direction: tc.Direction,
		ClientID:  sql.NullString{String: client, Valid: true},
		Limit:     tc.Limit,
	}
	if err == sql.ErrNoRows {
		//create new
		err = s.createTransCat(t)
	} else if tc.Limit.Valid {
		err = s.updateTransCatLimit(t)
	} else if !tc.Active.Bool && tc.Active.Valid {
		err = s.disableTransCat(t)
	}

	tc.ClientID = sql.NullString{String: client, Valid: true}
	fmt.Println(s.EditDocCategory(context.Background(), tc))

	return
}

func (s *Repo) PostDoc(ctx context.Context, time time.Time, category string, amount int, description string, msg_id string, direction int, client string) (err error) {
	doc := &DBDocument{
		Time:        &time,
		Category:    &category,
		Amount:      &amount,
		Description: &description,
		MsgID:       &msg_id,
		ClientID:    &client,
		Direction:   &direction,
	}

	return s.postDocument(doc)
}

func (s *Repo) DeleteDoc(msg_id string, client string) (err error) {
	tx := s.MustBegin()

	tx.MustExec("DELETE FROM public.document WHERE tg_msg_id = $1 and client_id = $2;", msg_id, client)

	return tx.Commit()
}

func (s *Repo) ClearUserHistory(username string) (err error) {
	tx := s.MustBegin()

	tx.MustExec("DELETE FROM public.document WHERE client_id = $1;", username)
	tx.MustExec("UPDATE public.trans_category SET active = false WHERE client_id = $1;", username)

	return tx.Commit()
}

func (s *Repo) ImportDocs(data []byte, client string) (err error) {
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
			ClientID:    &client,
			Direction:   v.Direction,
		})
		if err != nil {
			return
		}
	}

	tx := s.MustBegin()
	sql := "INSERT INTO trans_category(trans_cat, direction, client_id)" +
		" SELECT distinct trans_cat, direction, client_id FROM document WHERE client_id = $1"
	tx.MustExec(sql, client)

	return tx.Commit()
}

func (s *Repo) Export(client string) (rslt []byte, err error) {
	data, err := s.Postgres.Query("SELECT trans_date, trans_cat, trans_amount, comment, case direction when -1 then 'debit' when 1 then 'credit' else 'other' end as direction"+
		" FROM base.public.document WHERE client_id = $1 ORDER BY 1 DESC", client)
	if err != nil {
		return rslt, err
	}

	expDoc := entity.DocumentExport{}
	var (
		expDocs     []entity.DocumentExport
		description sql.NullString
	)

	for data.Next() {
		err = data.Scan(&expDoc.Time, &expDoc.Category, &expDoc.Amount, &description, &expDoc.Direction)
		if err != nil {
			return rslt, err
		}
		expDoc.Description = description.String
		expDocs = append(expDocs, expDoc)
	}
	return json.Marshal(expDocs)

}
