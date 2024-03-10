package repository

import (
	"context"
	"database/sql"
	"errors"
	"finance-tg-bot/internal/entity"
	"finance-tg-bot/pkg/ydb"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

type TransCat struct {
	Category  sql.NullString `db:"trans_cat"`
	Direction sql.NullInt16  `db:"direction"`
	ClientID  sql.NullString `db:"client_id"`
	Active    sql.NullBool   `db:"active"`
	Limit     sql.NullInt64  `db:"trans_limit"`
}

type DBDocument struct {
	TransDate   sql.NullTime   `db:"trans_date"   json:"trans_date"`
	Category    sql.NullString `db:"trans_cat"    json:"trans_cat"`
	Amount      sql.NullInt64  `db:"trans_amount" json:"trans_amount"`
	Description sql.NullString `db:"comment"      json:"comment"`
	MsgID       sql.NullString `db:"tg_msg_id"`
	ChatID      sql.NullString `db:"tg_chat_id"`
	ClientID    sql.NullString `db:"client_id"`
	Direction   sql.NullInt16  `db:"direction"    json:"direction"`
}

func PostDocument(db ydb.Ydb, ctx context.Context, doc *DBDocument) (err error) {
	query := `	DECLARE $trans_date   AS Datetime;
				DECLARE $trans_cat    AS String;	
				DECLARE $trans_amount AS Int64;
				DECLARE $comment      AS String;
				DECLARE $tg_msg_id    AS String;
				DECLARE $tg_chat_id   AS String;
				DECLARE $client_id    AS String;
				DECLARE $direction    AS Int8;`
	if doc.Direction.Valid {
		query = query + `
 UPSERT INTO document ( trans_date, trans_cat, trans_amount, comment, tg_msg_id, tg_chat_id, client_id, direction )
 VALUES ( $trans_date, $trans_cat, $trans_amount, $comment, $tg_msg_id, $tg_chat_id, $client_id, $direction );`
	} else {
		query = query + `
UPSERT INTO document ( trans_date, trans_cat, trans_amount, comment, tg_msg_id, tg_chat_id, client_id, direction )
SELECT $trans_date   as trans_date
     , trans_cat     as trans_cat
	 , $trans_amount as trans_amount
	 , $comment      as comment
	 , $tg_msg_id    as tg_msg_id
	 , $tg_chat_id   as tg_chat_id
	 , client_id     as client_id
	 , direction     as direction
  FROM doc_category
 WHERE trans_cat = $trans_cat
   AND client_id = $client_id
   AND active;`
	}

	err = db.Table().DoTx( // Do retry operation on errors with best effort
		ctx, // context manages exiting from Do
		func(ctx context.Context, tx table.TransactionActor) (err error) { // retry operation
			res, err := tx.Execute(
				ctx,
				query,
				table.NewQueryParameters(
					table.ValueParam("$trans_date", types.DatetimeValueFromTime(time.Now())),
					table.ValueParam("$trans_cat", types.BytesValueFromString(doc.Category.String)),
					table.ValueParam("$trans_amount", types.Int64Value(doc.Amount.Int64)),
					table.ValueParam("$comment", types.BytesValueFromString(doc.Description.String)),
					table.ValueParam("$tg_msg_id", types.BytesValueFromString(doc.MsgID.String)),
					table.ValueParam("$tg_chat_id", types.BytesValueFromString(doc.ChatID.String)),
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

func DeleteDocument(db ydb.Ydb, ctx context.Context, doc *DBDocument) (err error) {
	if !doc.MsgID.Valid || !doc.ClientID.Valid {
		return errors.New("not enough input params to delete document")
	}

	err = db.Table().DoTx( // Do retry operation on errors with best effort
		ctx, // context manages exiting from Do
		func(ctx context.Context, tx table.TransactionActor) (err error) { // retry operation
			res, err := tx.Execute(
				ctx, `DECLARE $tg_msg_id    AS String;
				      DECLARE $tg_chat_id   AS String;
				      DECLARE $client_id    AS String;
 DELETE FROM document
  WHERE tg_msg_id = $tg_msg_id
    AND client_id = $client_id;`,
				table.NewQueryParameters(
					table.ValueParam("$tg_msg_id", types.BytesValueFromString(doc.MsgID.String)),
					table.ValueParam("$tg_chat_id", types.BytesValueFromString(doc.ChatID.String)),
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

func GetDocumentCategories(db ydb.Ydb, ctx context.Context, username, limit string) (cats []entity.TransCatLimit, err error) {
	query := `DECLARE $client_id      AS String;
	          DECLARE $month_interval AS Datetime;
			  DECLARE $date_interval  AS Datetime;`
	if limit == "setting" {
		query = query + `
SELECT trans_cat, direction, trans_limit
  FROM doc_category
 WHERE client_id = $client_id
   AND active;`
	} else {
		query = query + `
SELECT dc.trans_cat        AS trans_cat
     , dc.direction        AS direction
	 , count(d.trans_date) AS cnt
     , dc.trans_limit
	 - sum(case when d.trans_date >= $month_interval then d.trans_amount
				  else 0 end) AS trans_limit
  FROM doc_category dc
  LEFT JOIN document d on (d.trans_cat = dc.trans_cat
					   and d.client_id = dc.client_id)
 WHERE dc.client_id = $client_id
   AND (d.trans_date is null
	 OR d.trans_date > $date_interval)
   AND dc.active
 GROUP BY dc.trans_cat, dc.direction, dc.trans_limit
 ORDER BY cnt desc;`
	}

	err = db.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		t := time.Now()
		tcl := &entity.TransCatLimit{}
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			query,
			table.NewQueryParameters(
				table.ValueParam("$client_id", types.BytesValueFromString(username)),
				table.ValueParam("$month_interval",
					types.DatetimeValueFromTime(time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()))),
				table.ValueParam("$date_interval", types.DatetimeValueFromTime(t.AddDate(0, -3, 0))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		if err = res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			err = res.ScanNamed(
				named.Optional("trans_cat", &tcl.Category),
				named.Optional("direction", &tcl.Direction),
				named.Optional("trans_limit", &tcl.Limit),
			)
			if err != nil {
				return err
			}
			cats = append(cats, *tcl)
		}
		return res.Err() // for driver retry if not nil
	})

	return cats, err
}

func EditCategory(db ydb.Ydb, ctx context.Context, cat *TransCat) (err error) {
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
		      ,date_to = $date_to
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
	
			UPSERT INTO doc_category ( trans_cat, direction, client_id, active, date_to, trans_limit )
			VALUES ( $trans_cat, $direction, $client_id, $active, $date_to_max, $trans_limit );`
	} else {
		query = query + `
		UPSERT INTO doc_category ( trans_cat, direction, client_id, date_to, active )
		VALUES ( $trans_cat, $direction, $client_id, $date_to_max, $active );`
	}

	err = db.Table().DoTx( // Do retry operation on errors with best effort
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

func GetDocumentSubCategories(db ydb.Ydb, ctx context.Context, username, trans_cat string) (subcats []string, err error) {
	query := `DECLARE $client_id      AS String;
			  DECLARE $trans_cat 	  AS String;
			  DECLARE $date_interval  AS Datetime;
SELECT comment, count(*) AS cnt
  FROM document
 WHERE comment != ''
   AND client_id = $client_id
   AND trans_cat = $trans_cat
   AND trans_date > $date_interval
 GROUP BY comment
 ORDER BY cnt DESC
 LIMIT 20;`
	err = db.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		t := time.Now()
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			query,
			table.NewQueryParameters(
				table.ValueParam("$client_id", types.BytesValueFromString(username)),
				table.ValueParam("$trans_cat", types.BytesValueFromString(trans_cat)),
				table.ValueParam("$date_interval", types.DatetimeValueFromTime(t.AddDate(0, -3, 0))),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		if err = res.NextResultSetErr(ctx); err != nil {
			return err
		}
		var subcat sql.NullString
		for res.NextRow() {
			err = res.ScanNamed(named.Optional("comment", &subcat))

			if err != nil {
				return err
			}
			subcats = append(subcats, subcat.String)
		}
		return res.Err() // for driver retry if not nil
	})

	return
}
