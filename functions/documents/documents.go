package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
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
	prefix string
}

func connectDB(ctx context.Context, dsn, saPath, prefix string) (*Ydb, error) {
	var opt ydb.Option
	if saPath == "" {
		// auth inside cloud (virual machine or yandex function)
		opt = yc.WithMetadataCredentials()
	} else {
		// auth from service account key file
		opt = yc.WithServiceAccountKeyFileCredentials(saPath)
	}
	nativeDriver, err := ydb.Open(ctx, dsn, yc.WithInternalCA(), opt)
	if prefix != "" {
		prefix = path.Join(nativeDriver.Name(), prefix)
	}

	return &Ydb{Driver: nativeDriver, prefix: prefix}, err
}

type TransCatLimit struct {
	Category  *string `json:"trans_cat"`
	Direction *int8   `json:"direction"`
	UserId    int     `json:"client_id,omitempty"`
	Active    bool    `json:"active,omitempty"`
	Limit     *int64  `json:"trans_limit"`
	Balance   *int64  `json:"trans_balance"`
}

type DBDocument struct {
	RecTime     time.Time `json:"rec_time"`
	TransDate   time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int64     `json:"trans_amount"`
	Description string    `json:"comment"`
	MsgID       string    `json:"msg_id"`
	ChatID      string    `json:"chat_id"`
	UserId      int       `json:"client_id"`
	Direction   int8      `json:"direction"`
}

func (r *Ydb) PostDocument(ctx context.Context, doc *DBDocument) (err error) {
	//preformat
	doc.Description = strings.ToLower(strings.TrimSpace(doc.Description))
	//
	query := `	DECLARE $rec_time     AS Timestamp;
				DECLARE $trans_date   AS Datetime;
				DECLARE $trans_cat    AS String;	
				DECLARE $trans_amount AS Int64;
				DECLARE $comment      AS String;
				DECLARE $msg_id       AS String;
				DECLARE $chat_id      AS String;
				DECLARE $client_id    AS Uint64;
				DECLARE $direction    AS Int8;
 UPSERT INTO doc ( rec_time, trans_date, trans_cat, trans_amount, comment, msg_id, chat_id, client_id, direction )
 VALUES ( $rec_time, $trans_date, $trans_cat, $trans_amount, $comment, $msg_id, $chat_id, $client_id, $direction );`

	if r.prefix != "" {
		query = fmt.Sprintf(`PRAGMA TablePathPrefix("%s"); %s`, r.prefix, query)
	}

	err = r.Table().DoTx( // Do retry operation on errors with best effort
		ctx, // context manages exiting from Do
		func(ctx context.Context, tx table.TransactionActor) (err error) { // retry operation
			res, err := tx.Execute(
				ctx,
				query,
				table.NewQueryParameters(
					table.ValueParam("$rec_time", types.TimestampValueFromTime(doc.RecTime)),
					table.ValueParam("$trans_date", types.DatetimeValueFromTime(doc.TransDate)),
					table.ValueParam("$trans_cat", types.BytesValueFromString(doc.Category)),
					table.ValueParam("$trans_amount", types.Int64Value(doc.Amount)),
					table.ValueParam("$comment", types.BytesValueFromString(doc.Description)),
					table.ValueParam("$msg_id", types.BytesValueFromString(doc.MsgID)),
					table.ValueParam("$chat_id", types.BytesValueFromString(doc.ChatID)),
					table.ValueParam("$client_id", types.Uint64Value(uint64(doc.UserId))),
					table.ValueParam("$direction", types.Int8Value(doc.Direction)),
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
	if doc.MsgID == "" || doc.UserId == 0 {
		return errors.New("not enough input params to delete document")
	}

	query := `DECLARE $msg_id      AS String;
              DECLARE $chat_id     AS String;
              DECLARE $client_id   AS Uint64;
DELETE FROM doc
 WHERE msg_id = $msg_id
   AND chat_id = $chat_id
   AND client_id = $client_id;`

	if r.prefix != "" {
		query = fmt.Sprintf(`PRAGMA TablePathPrefix("%s"); %s`, r.prefix, query)
	}

	err = r.Table().DoTx( // Do retry operation on errors with best effort
		ctx, // context manages exiting from Do
		func(ctx context.Context, tx table.TransactionActor) (err error) { // retry operation
			res, err := tx.Execute(
				ctx,
				query,
				table.NewQueryParameters(
					table.ValueParam("$msg_id", types.BytesValueFromString(doc.MsgID)),
					table.ValueParam("$chat_id", types.BytesValueFromString(doc.ChatID)),
					table.ValueParam("$client_id", types.Uint64Value(uint64(doc.UserId))),
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

func (r *Ydb) GetDocumentCategories(ctx context.Context, user_id int) (cats []TransCatLimit, err error) {
	query := `
DECLARE $client_id      AS Uint64;
DECLARE $month_interval AS Datetime;
DECLARE $date_interval  AS Datetime;

SELECT dc.trans_cat        AS trans_cat
     , dc.direction        AS direction
	 , count(d.trans_date) AS cnt
	 , dc.trans_limit	   AS trans_limit
     , dc.trans_limit
	 - sum(case when d.trans_date >= $month_interval then d.trans_amount
				  else 0 end) AS trans_balance
  FROM doc_cat dc
  LEFT JOIN doc d on (d.trans_cat = dc.trans_cat
				  and d.client_id = dc.client_id)
 WHERE dc.active
   AND dc.client_id = $client_id
   AND (d.trans_date is null
	 OR d.trans_date > $date_interval)
 GROUP BY dc.trans_cat, dc.direction, dc.trans_limit
 ORDER BY cnt desc;`

	if r.prefix != "" {
		query = fmt.Sprintf(`PRAGMA TablePathPrefix("%s"); %s`, r.prefix, query)
	}

	err = r.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		t := time.Now()
		tcl := &TransCatLimit{}
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			query,
			table.NewQueryParameters(
				table.ValueParam("$client_id", types.Uint64Value(uint64(user_id))),
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

func (r *Ydb) EditCategory(ctx context.Context, cat *TransCatLimit) (err error) {
	//preformat
	*cat.Category = strings.ToLower(strings.TrimSpace(*cat.Category))
	//
	query := `	DECLARE $trans_cat    AS String;
				DECLARE $direction    AS Int8;
				DECLARE $client_id    AS Uint64;
				DECLARE $active       AS Bool;
				DECLARE $date_to      AS Datetime;
				DECLARE $trans_limit  AS Int64;
				DECLARE $date_to_max  AS Datetime;`
	if !cat.Active {
		query = query + `
		UPDATE doc_cat
		   SET active = false
		 WHERE active
		   AND date_to = $date_to_max
		   AND trans_cat = $trans_cat
		   AND client_id = $client_id;`
	} else if *cat.Limit > 0 {
		query = query + `
		UPSERT INTO doc_cat ( trans_cat, direction, client_id, active, date_to, trans_limit )
		SELECT trans_cat, direction, client_id
			 , false as active, $date_to as date_to
			 , trans_limit
		  FROM doc_cat
		 WHERE active
		   AND date_to = $date_to_max
		   AND trans_cat = $trans_cat
		   AND client_id = $client_id;

		UPDATE doc_cat
		   SET trans_limit = $trans_limit
		 WHERE active
		   AND date_to = $date_to_max
		   AND trans_cat = $trans_cat
		   AND client_id = $client_id;`
	} else {
		query = query + `
		UPSERT INTO doc_cat ( trans_cat, direction, client_id, date_to, active )
		VALUES ( $trans_cat, $direction, $client_id, $date_to_max, $active );`
	}

	if r.prefix != "" {
		query = fmt.Sprintf(`PRAGMA TablePathPrefix("%s"); %s`, r.prefix, query)
	}

	err = r.Table().DoTx( // Do retry operation on errors with best effort
		ctx, // context manages exiting from Do
		func(ctx context.Context, tx table.TransactionActor) (err error) { // retry operation
			t, _ := time.Parse("2006-01-02", "2100-01-01")
			res, err := tx.Execute(
				ctx,
				query,
				table.NewQueryParameters(
					table.ValueParam("$trans_cat", types.BytesValueFromString(*cat.Category)),
					table.ValueParam("$direction", types.Int8Value(*cat.Direction)),
					table.ValueParam("$client_id", types.Uint64Value(uint64(cat.UserId))),
					table.ValueParam("$active", types.BoolValue(cat.Active)),
					table.ValueParam("$date_to", types.DatetimeValueFromTime(time.Now())),
					table.ValueParam("$trans_limit", types.Int64Value(int64(*cat.Limit))),
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

func (r *Ydb) GetDocumentSubCategories(ctx context.Context, user_id int, trans_cat string) (subcats []string, err error) {
	query := `DECLARE $client_id      AS Uint64;
			  DECLARE $trans_cat 	  AS String;
			  DECLARE $date_interval  AS Timestamp;
SELECT d.comment as comment
     , count(*) AS cnt
  FROM doc d
 WHERE d.comment != ''
   AND d.trans_cat = $trans_cat
   AND d.rec_time > $date_interval
   AND d.client_id = $client_id
 GROUP BY d.comment
 ORDER BY cnt DESC
 LIMIT 25;`

	if r.prefix != "" {
		query = fmt.Sprintf(`PRAGMA TablePathPrefix("%s"); %s`, r.prefix, query)
	}

	err = r.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		t := time.Now()
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			query,
			table.NewQueryParameters(
				table.ValueParam("$client_id", types.Uint64Value(uint64(user_id))),
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
