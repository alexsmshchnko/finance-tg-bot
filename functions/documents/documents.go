package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	yc "github.com/ydb-platform/ydb-go-yc"
)

type Ydb struct {
	*ydb.Driver
}

var db *Ydb

func connectDB(ctx context.Context, dsn, saPath string) (*Ydb, error) {
	var opt ydb.Option
	if saPath == "" {
		// auth inside cloud (virual machine or yandex function)
		opt = yc.WithMetadataCredentials()
	} else {
		// auth from service account key file
		opt = yc.WithServiceAccountKeyFileCredentials(saPath)
	}
	nativeDriver, err := ydb.Open(ctx, dsn, yc.WithInternalCA(), opt)

	return &Ydb{Driver: nativeDriver}, err
}

type TransCat struct {
	Category  sql.NullString `db:"trans_cat"`
	Direction sql.NullInt16  `db:"direction"`
	ClientID  sql.NullString `db:"client_id"`
	Active    sql.NullBool   `db:"active"`
	Limit     sql.NullInt64  `db:"trans_limit"`
}

type TransCatLimit struct {
	Category  sql.NullString `db:"trans_cat"`
	Direction sql.NullInt16  `db:"direction"`
	ClientID  sql.NullString `db:"client_id"`
	Active    sql.NullBool   `db:"active"`
	Limit     sql.NullInt64  `db:"trans_limit"`
	Balance   sql.NullInt64  `db:"trans_balance"`
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

func (r *Ydb) PostDocument(ctx context.Context, doc *DBDocument) (err error) {
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

	err = r.Table().DoTx( // Do retry operation on errors with best effort
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

func (r *Ydb) DeleteDocument(ctx context.Context, doc *DBDocument) (err error) {
	if !doc.MsgID.Valid || !doc.ClientID.Valid {
		return errors.New("not enough input params to delete document")
	}

	err = r.Table().DoTx( // Do retry operation on errors with best effort
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

func (r *Ydb) GetDocumentCategories(ctx context.Context, username, limit string) (cats []TransCatLimit, err error) {
	query := `DECLARE $client_id      AS String;
	          DECLARE $month_interval AS Datetime;
			  DECLARE $date_interval  AS Datetime;
SELECT dc.trans_cat        AS trans_cat
     , dc.direction        AS direction
	 , count(d.trans_date) AS cnt
	 , dc.trans_limit	   AS trans_limit
     , dc.trans_limit
	 - sum(case when d.trans_date >= $month_interval then d.trans_amount
				  else 0 end) AS trans_balance
  FROM doc_category dc
 INNER JOIN client c on (c.username = dc.client_id)
  LEFT JOIN doc d on (d.trans_cat = dc.trans_cat
				  and d.client_id = c.id)
 WHERE dc.active
   AND (d.trans_date is null
	 OR d.trans_date > $date_interval)
   AND c.username = $client_id
 GROUP BY dc.trans_cat, dc.direction, dc.trans_limit
 ORDER BY cnt desc;`

	err = r.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		t := time.Now()
		tcl := &TransCatLimit{}
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
				named.Optional("trans_balance", &tcl.Balance),
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

func (r *Ydb) EditCategory(ctx context.Context, cat *TransCat) (err error) {
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

	err = r.Table().DoTx( // Do retry operation on errors with best effort
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

func (r *Ydb) GetDocumentSubCategories(ctx context.Context, username, trans_cat string) (subcats []string, err error) {
	query := `DECLARE $client_id      AS String;
			  DECLARE $trans_cat 	  AS String;
			  DECLARE $date_interval  AS Timestamp;
SELECT d.comment as comment
     , count(*) AS cnt
  FROM doc d
 INNER JOIN client c ON (c.id = d.client_id)
 WHERE d.comment != ''
   AND d.trans_cat = $trans_cat
   AND d.rec_time > $date_interval
   AND c.username = $client_id
   AND c.is_active
 GROUP BY d.comment
 ORDER BY cnt DESC
 LIMIT 20;`
	err = r.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		t := time.Now()
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			query,
			table.NewQueryParameters(
				table.ValueParam("$client_id", types.BytesValueFromString(username)),
				table.ValueParam("$trans_cat", types.BytesValueFromString(trans_cat)),
				table.ValueParam("$date_interval", types.TimestampValueFromTime(t.AddDate(0, -3, 0))),
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
