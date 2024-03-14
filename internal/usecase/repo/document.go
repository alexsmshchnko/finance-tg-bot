package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"finance-tg-bot/internal/entity"
	"finance-tg-bot/pkg/repository"
	"finance-tg-bot/pkg/ydb"
)

type Repo struct {
	// *postgres.Postgres
	*ydb.Ydb
}

func New( //pg *postgres.Postgres,
	ydb *ydb.Ydb) *Repo {
	return &Repo{Ydb: ydb}
}

// type DBDocument struct {
// 	ID          *int64     `db:"id"`
// 	Time        *time.Time `db:"trans_date"   json:"trans_date"`
// 	Category    *string    `db:"trans_cat"    json:"trans_cat"`
// 	Amount      *int       `db:"trans_amount" json:"trans_amount"`
// 	Description *string    `db:"comment"      json:"comment"`
// 	MsgID       *string    `db:"tg_msg_id"`
// 	ClientID    *string    `db:"client_id"`
// 	Direction   *int       `db:"direction"    json:"direction"`
// }

// type TransCat struct {
// 	ID        sql.NullInt64  `db:"id"`
// 	Category  sql.NullString `db:"trans_cat"`
// 	Direction sql.NullInt16  `db:"direction"`
// 	ClientID  sql.NullString `db:"client_id"`
// 	Active    sql.NullBool   `db:"active"`
// 	Limit     sql.NullInt64  `db:"trans_limit"`
// }

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

func (s *Repo) GetCategories(ctx context.Context, username, limit string) (cat []entity.TransCatLimit, err error) {
	return repository.GetDocumentCategories(*s.Ydb, ctx, username, limit)
}

func (s *Repo) GetCats(ctx context.Context, username, limit string) (cat []entity.TransCatLimit, err error) {
	// var sql string
	// switch limit {
	// case "setting":
	// 	sql = `
	// 	select tc.trans_cat, tc.direction, tc.trans_limit
	// 	  from public.trans_category tc
	// 	  left join public.document d on (tc.trans_cat = d.trans_cat
	// 								  and tc.client_id = d.client_id)
	// 	 where tc.active = true
	// 	   and tc.client_id = $1
	// 	 group by tc.trans_cat, tc.direction, tc.trans_limit
	// 	 order by count(*) desc`
	// case "balance":
	// 	sql = `
	// 	select tc.trans_cat, tc.direction
	// 	      ,tc.trans_limit - sum(case when d.trans_date >= date_trunc('month', current_date) then d.trans_amount else 0 end) as trans_limit
	//       from public.trans_category tc
	//       left join public.document d on (d.trans_cat = tc.trans_cat
	// 							      and d.client_id = tc.client_id
	// 							      and d.trans_date between date_trunc('month', current_date - interval '3' month)
	// 								                       and date_trunc('day', current_date + interval '1' day) - interval '1' second)
	//      where tc.active = true
	//        and tc.client_id = $1
	//      group by tc.trans_cat, tc.direction, tc.trans_limit
	//      order by count(d.*) desc`
	// }
	// err = s.Select(&cat, sql, username)
	return
}

func (s *Repo) GetSubCategories(ctx context.Context, username, trans_cat string) (cat []string, err error) {
	return repository.GetDocumentSubCategories(*s.Ydb, ctx, username, trans_cat)
}

func (s *Repo) EditCategory(ctx context.Context, tc entity.TransCatLimit, client string) (err error) {
	dbTCL := &repository.TransCat{
		Category:  tc.Category,
		Direction: tc.Direction,
		ClientID:  sql.NullString{String: client, Valid: true},
		Active:    tc.Active,
	}
	if tc.Limit.Valid {
		dbTCL.Limit = tc.Limit
	}

	return repository.EditCategory(*s.Ydb, ctx, dbTCL)
}

func (s *Repo) ClearUserHistory(username string) (err error) {
	// tx := s.MustBegin()

	// tx.MustExec("DELETE FROM public.document WHERE client_id = $1;", username)
	// tx.MustExec("UPDATE public.trans_category SET active = false WHERE client_id = $1;", username)

	// return tx.Commit()
	return
}

func (s *Repo) ImportDocs(data []byte, client string) (err error) {
	// var docs []DBDocument

	// err = json.Unmarshal(data, &docs)
	// if err != nil {
	// 	return err
	// }

	// for _, v := range docs {
	// 	s.postDocument(&DBDocument{
	// 		Time:        v.Time,
	// 		Category:    v.Category,
	// 		Amount:      v.Amount,
	// 		Description: v.Description,
	// 		ClientID:    &client,
	// 		Direction:   v.Direction,
	// 	})
	// }

	// tx := s.MustBegin()
	// sql := "INSERT INTO trans_category(trans_cat, direction, client_id)" +
	// 	" SELECT distinct trans_cat, direction, client_id FROM document WHERE client_id = $1"
	// tx.MustExec(sql, client)

	// return tx.Commit()
	return errors.New("not working")
}

func (s *Repo) Export(client string) (rslt []byte, err error) {
	// data, err := s.Postgres.Query(`
	// SELECT trans_date,
	// trans_cat,
	// trans_amount,
	// comment,
	// case direction when -1 then 'debit' when 1 then 'credit' else 'other' end as direction,
	// tg_msg_id
	//   FROM base.public.document WHERE client_id = $1 ORDER BY 1 DESC`, client)
	// if err != nil {
	// 	return rslt, err
	// }

	// expDoc := entity.DocumentExport{}
	// var (
	// 	expDocs     []entity.DocumentExport
	// 	description sql.NullString
	// 	msgId       sql.NullString
	// )

	// for data.Next() {
	// 	err = data.Scan(&expDoc.Time, &expDoc.Category, &expDoc.Amount, &description, &expDoc.Direction, &msgId)
	// 	if err != nil {
	// 		return rslt, err
	// 	}
	// 	expDoc.Description = description.String
	// 	expDoc.MsgID = msgId.String
	// 	expDoc.ClientID = client
	// 	expDocs = append(expDocs, expDoc)
	// }
	// return json.Marshal(expDocs)

	return rslt, errors.New("not working")
}
