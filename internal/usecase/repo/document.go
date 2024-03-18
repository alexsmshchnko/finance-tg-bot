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

type categories []struct {
	category    string
	tcl         entity.TransCatLimit
	subcategory []string
}

type Repo struct {
	// *postgres.Postgres
	*ydb.Ydb
	cache map[string]categories
}

func (s *Repo) clearCache(username string) {
	delete(s.cache, username)
}

func New(ydb *ydb.Ydb) *Repo {
	return &Repo{
		Ydb:   ydb,
		cache: make(map[string]categories),
	}
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
	s.clearCache(doc.ClientID)
	err = repository.PostDocument(*s.Ydb, ctx, dbdoc)
	go s.GetCategories(ctx, doc.ClientID, "")
	return
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
	if _, ok := s.cache[username]; !ok {
		res, err := repository.GetDocumentCategories(*s.Ydb, ctx, username, limit)
		if err != nil {
			return cat, err
		}
		s.cache[username] = make(categories, len(res))
		for i, v := range res {
			s.cache[username][i] = struct {
				category    string
				tcl         entity.TransCatLimit
				subcategory []string
			}{category: v.Category.String, tcl: v}
		}
	}

	cat = make([]entity.TransCatLimit, len(s.cache[username]))
	for i, v := range s.cache[username] {
		cat[i] = v.tcl
	}
	return cat, err
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
	if _, ok := s.cache[username]; !ok { //cache should present after GetCategories request
		s.GetCategories(ctx, username, "") //in case we missed smth
	}

	var j int
	for i, v := range s.cache[username] { //search cache
		if v.category == trans_cat && len(v.subcategory) > 0 {
			return s.cache[username][i].subcategory, err //return if found
		}
		if v.category == trans_cat {
			j = i
			break
		}
	}

	res, err := repository.GetDocumentSubCategories(*s.Ydb, ctx, username, trans_cat)
	if err != nil {
		return cat, err
	}

	//create cache
	s.cache[username][j].subcategory = res

	return s.cache[username][j].subcategory, err
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

	s.clearCache(client)
	err = repository.EditCategory(*s.Ydb, ctx, dbTCL)
	go s.GetCategories(ctx, client, "")
	return
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
