package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"finance-tg-bot/internal/entity"
	repPkg "finance-tg-bot/pkg/repository"
)

type categories []struct {
	category    string
	tcl         entity.TransCatLimit
	subcategory []string
}

type Repo struct {
	repo  repPkg.DocProcessor
	cache map[string]categories
}

func New(repPkg repPkg.DocProcessor) *Repo {
	return &Repo{
		repo:  repPkg,
		cache: make(map[string]categories),
	}
}

func (r *Repo) clearCache(username string) {
	delete(r.cache, username)
}

func (r *Repo) PostDocument(ctx context.Context, doc *entity.Document) (err error) {
	dbdoc := &repPkg.DBDocument{
		TransDate:   sql.NullTime{Time: time.Time{}, Valid: false},
		Category:    sql.NullString{String: doc.Category, Valid: true},
		Amount:      sql.NullInt64{Int64: doc.Amount, Valid: true},
		Description: sql.NullString{String: doc.Description, Valid: true},
		MsgID:       sql.NullString{String: doc.MsgID, Valid: true},
		ChatID:      sql.NullString{String: doc.ChatID, Valid: true},
		ClientID:    sql.NullString{String: doc.ClientID, Valid: true},
		Direction:   sql.NullInt16{Int16: 0, Valid: false},
	}
	r.clearCache(doc.ClientID)
	err = r.repo.PostDocument(ctx, dbdoc)
	go r.GetCategories(ctx, doc.ClientID, "")
	return
}

func (r *Repo) DeleteDocument(ctx context.Context, doc *entity.Document) (err error) {
	dbdoc := &repPkg.DBDocument{
		MsgID:    sql.NullString{String: doc.MsgID, Valid: true},
		ChatID:   sql.NullString{String: doc.ChatID, Valid: true},
		ClientID: sql.NullString{String: doc.ClientID, Valid: true},
	}
	return r.repo.DeleteDocument(ctx, dbdoc)
}

func (r *Repo) GetCategories(ctx context.Context, username, limit string) (cat []entity.TransCatLimit, err error) {
	if _, ok := r.cache[username]; !ok {
		res, err := r.repo.GetDocumentCategories(ctx, username, limit)
		if err != nil {
			return cat, err
		}
		r.cache[username] = make(categories, len(res))
		for i, v := range res {
			r.cache[username][i] = struct {
				category    string
				tcl         entity.TransCatLimit
				subcategory []string
			}{category: v.Category.String, tcl: v}
		}
	}

	cat = make([]entity.TransCatLimit, len(r.cache[username]))
	for i, v := range r.cache[username] {
		cat[i] = v.tcl
	}
	return cat, err
}

func (r *Repo) GetSubCategories(ctx context.Context, username, trans_cat string) (cat []string, err error) {
	if _, ok := r.cache[username]; !ok { //cache should present after GetCategories request
		r.GetCategories(ctx, username, "") //in case we missed smth
	}

	var j int
	for i, v := range r.cache[username] { //search cache
		if v.category == trans_cat && len(v.subcategory) > 0 {
			return r.cache[username][i].subcategory, err //return if found
		}
		if v.category == trans_cat {
			j = i
			break
		}
	}

	res, err := r.repo.GetDocumentSubCategories(ctx, username, trans_cat)
	if err != nil {
		return cat, err
	}

	//create cache
	r.cache[username][j].subcategory = res

	return r.cache[username][j].subcategory, err
}

func (r *Repo) EditCategory(ctx context.Context, tc entity.TransCatLimit, client string) (err error) {
	dbTCL := &repPkg.TransCat{
		Category:  tc.Category,
		Direction: tc.Direction,
		ClientID:  sql.NullString{String: client, Valid: true},
		Active:    tc.Active,
	}
	if tc.Limit.Valid {
		dbTCL.Limit = tc.Limit
	}

	r.clearCache(client)
	err = r.repo.EditCategory(ctx, dbTCL)
	go r.GetCategories(ctx, client, "")
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
