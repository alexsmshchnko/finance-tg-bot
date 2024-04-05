package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"finance-tg-bot/internal/entity"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

type DocProcessor interface {
	PostDocument(ctx context.Context, doc *DBDocument) (err error)
	DeleteDocument(ctx context.Context, doc *DBDocument) (err error)
	GetDocumentCategories(ctx context.Context, username, limit string) (cats []entity.TransCatLimit, err error)
	EditCategory(ctx context.Context, cat *TransCat) (err error)
	GetDocumentSubCategories(ctx context.Context, username, trans_cat string) (subcats []string, err error)
}

type TransCat struct {
	Category  sql.NullString `db:"trans_cat"`
	Direction sql.NullInt16  `db:"direction"`
	ClientID  sql.NullString `db:"client_id"`
	Active    sql.NullBool   `db:"active"`
	Limit     sql.NullInt64  `db:"trans_limit"`
}

type DBDocument struct {
	RecDate     sql.NullTime   `db:"rec_time"`
	TransDate   sql.NullTime   `db:"trans_date"   json:"trans_date"`
	Category    sql.NullString `db:"trans_cat"    json:"trans_cat"`
	Amount      sql.NullInt64  `db:"trans_amount" json:"trans_amount"`
	Description sql.NullString `db:"comment"      json:"comment"`
	MsgID       sql.NullString `db:"msg_id"`
	ChatID      sql.NullString `db:"chat_id"`
	ClientID    sql.NullString `db:"client_id"`
	Direction   sql.NullInt16  `db:"direction"    json:"direction"`
}

func (r *Repository) PostDocument(ctx context.Context, doc *DBDocument) (err error) {
	//preformat
	doc.Description.String = strings.ToLower(strings.TrimSpace(doc.Description.String))
	//
	query := `	DECLARE $rec_time     AS Timestamp;
				DECLARE $trans_date   AS Datetime;
				DECLARE $trans_cat    AS String;	
				DECLARE $trans_amount AS Int64;
				DECLARE $comment      AS String;
				DECLARE $msg_id       AS String;
				DECLARE $chat_id      AS String;
				DECLARE $client_id    AS String;
				DECLARE $direction    AS Int8;`
	if doc.Direction.Valid {
		query = query + `
 UPSERT INTO doc ( rec_time, trans_date, trans_cat, trans_amount, comment, msg_id, chat_id, client_id, direction )
 SELECT $rec_time, $trans_date, $trans_cat, $trans_amount, $comment, $msg_id, $chat_id, c.id as client_id, $direction
   FROM client c
  WHERE c.is_active
    AND c.username = $client_id;`
	} else {
		query = query + `
UPSERT INTO doc ( rec_time, trans_date, trans_cat, trans_amount, comment, msg_id, chat_id, client_id, direction )
SELECT $rec_time     as rec_time
     , $trans_date   as trans_date
     , trans_cat     as trans_cat
	 , $trans_amount as trans_amount
	 , $comment      as comment
	 , $msg_id       as msg_id
	 , $chat_id      as chat_id
	 , c.id          as client_id
	 , direction     as direction
  FROM client c
 INNER JOIN doc_category dc ON (dc.client_id = c.username)
 WHERE c.is_active
   AND c.username = $client_id
   AND dc.active
   AND dc.trans_cat = $trans_cat;`
	}

	if !doc.TransDate.Valid {
		doc.TransDate.Time = time.Now()
		doc.TransDate.Valid = true
	}
	if !doc.RecDate.Valid {
		doc.RecDate.Time = time.Now()
		doc.RecDate.Valid = true
	}

	err = r.Ydb.Table().DoTx( // Do retry operation on errors with best effort
		ctx, // context manages exiting from Do
		func(ctx context.Context, tx table.TransactionActor) (err error) { // retry operation
			res, err := tx.Execute(
				ctx,
				query,
				table.NewQueryParameters(
					table.ValueParam("$rec_time", types.TimestampValueFromTime(doc.RecDate.Time)),
					table.ValueParam("$trans_date", types.DatetimeValueFromTime(doc.TransDate.Time)),
					table.ValueParam("$trans_cat", types.BytesValueFromString(doc.Category.String)),
					table.ValueParam("$trans_amount", types.Int64Value(doc.Amount.Int64)),
					table.ValueParam("$comment", types.BytesValueFromString(doc.Description.String)),
					table.ValueParam("$msg_id", types.BytesValueFromString(doc.MsgID.String)),
					table.ValueParam("$chat_id", types.BytesValueFromString(doc.ChatID.String)),
					table.ValueParam("$client_id", types.BytesValueFromString(doc.ClientID.String)),
					table.ValueParam("$direction", types.Int8Value(int8(doc.Direction.Int16))),
				),
			)
			if err != nil {
				return err
			}
			if err = res.Err(); err != nil {
				return err
			}
			return res.Close()
		}, table.WithIdempotent(),
	)

	return err
}

func (r *Repository) DeleteDocument(ctx context.Context, doc *DBDocument) (err error) {
	if !doc.MsgID.Valid || !doc.ClientID.Valid {
		return errors.New("not enough input params to delete document")
	}

	err = r.Ydb.Table().DoTx( // Do retry operation on errors with best effort
		ctx, // context manages exiting from Do
		func(ctx context.Context, tx table.TransactionActor) (err error) { // retry operation
			res, err := tx.Execute(
				ctx, `DECLARE $msg_id      AS String;
				      DECLARE $chat_id     AS String;
				      DECLARE $client_id   AS String;
 DELETE FROM doc
  WHERE msg_id = $msg_id
    AND client_id in (select c.id from client c where c.username = $client_id);`,
				table.NewQueryParameters(
					table.ValueParam("$msg_id", types.BytesValueFromString(doc.MsgID.String)),
					table.ValueParam("$chat_id", types.BytesValueFromString(doc.ChatID.String)),
					table.ValueParam("$client_id", types.BytesValueFromString(doc.ClientID.String)),
				),
			)
			if err != nil {
				return err
			}
			if err = res.Err(); err != nil {
				return err
			}
			return res.Close()
		}, table.WithIdempotent(),
	)

	return err
}

func (r *Repository) GetDocumentCategories(ctx context.Context, username, limit string) (cats []entity.TransCatLimit, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		r.ServiceDomain+"/category/"+username,
		nil,
	)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentCategories http.NewRequestWithContext", "err", err)
		return
	}

	req.Header.Add("Authorization", "Basic "+r.AuthToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentCategories client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.GetDocumentCategories response", "StatusCode", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentCategories io.ReadAll", "err", err)
		return
	}
	err = json.Unmarshal(body, &cats)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentCategories json.Unmarshal", "err", err)
	}

	return
}

func (r *Repository) EditCategory(ctx context.Context, cat *TransCat) (err error) {
	//preformat
	cat.Category.String = strings.ToLower(strings.TrimSpace(cat.Category.String))
	//
	query := `	DECLARE $trans_cat    AS String;
				DECLARE $direction    AS Int8;
				DECLARE $client_id    AS String;
				DECLARE $active       AS Bool;
				DECLARE $date_to      AS Datetime;
				DECLARE $trans_limit  AS Int64;
				DECLARE $date_to_max  AS Datetime;`
	if !cat.Active.Bool && cat.Active.Valid {
		query = query + `
		UPDATE doc_category
		   SET active = false
		 WHERE active
		   AND date_to = $date_to_max
		   AND trans_cat = $trans_cat
		   AND client_id = $client_id;`
	} else if cat.Limit.Valid {
		query = query + `
		UPSERT INTO doc_category ( trans_cat, direction, client_id, active, date_to, trans_limit )
		SELECT trans_cat, direction, client_id
			 , false as active, $date_to as date_to
			 , trans_limit
		  FROM doc_category
		 WHERE active
		   AND date_to = $date_to_max
		   AND trans_cat = $trans_cat
		   AND client_id = $client_id;

		UPDATE doc_category
		   SET trans_limit = $trans_limit
		 WHERE active
		   AND date_to = $date_to_max
		   AND trans_cat = $trans_cat
		   AND client_id = $client_id;`
	} else {
		query = query + `
		UPSERT INTO doc_category ( trans_cat, direction, client_id, date_to, active )
		VALUES ( $trans_cat, $direction, $client_id, $date_to_max, $active );`
	}

	err = r.Ydb.Table().DoTx( // Do retry operation on errors with best effort
		ctx, // context manages exiting from Do
		func(ctx context.Context, tx table.TransactionActor) (err error) { // retry operation
			t, _ := time.Parse("2006-01-02", "2100-01-01")
			res, err := tx.Execute(
				ctx,
				query,
				table.NewQueryParameters(
					table.ValueParam("$trans_cat", types.BytesValueFromString(cat.Category.String)),
					table.ValueParam("$direction", types.Int8Value(int8(cat.Direction.Int16))),
					table.ValueParam("$client_id", types.BytesValueFromString(cat.ClientID.String)),
					table.ValueParam("$active", types.BoolValue(cat.Active.Bool)),
					table.ValueParam("$date_to", types.DatetimeValueFromTime(time.Now())),
					table.ValueParam("$trans_limit", types.Int64Value(cat.Limit.Int64)),
					table.ValueParam("$date_to_max", types.DatetimeValueFromTime(t)),
				),
			)
			if err != nil {
				return err
			}
			if err = res.Err(); err != nil {
				return err
			}
			return res.Close()
		}, table.WithIdempotent(),
	)

	return err
}

func (r *Repository) GetDocumentSubCategories(ctx context.Context, username, trans_cat string) (subcats []string, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"OPTIONS",
		r.ServiceDomain+"/category/"+username+"/"+trans_cat,
		nil,
	)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentSubCategories http.NewRequestWithContext", "err", err)
		return
	}

	req.Header.Add("Authorization", "Basic "+r.AuthToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentSubCategories client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.GetDocumentSubCategories response", "StatusCode", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentSubCategories io.ReadAll", "err", err)
		return
	}
	err = json.Unmarshal(body, &subcats)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentSubCategories json.Unmarshal", "err", err)
	}

	return
}
