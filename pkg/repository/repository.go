package repository

import (
	"context"
	"database/sql"
	"finance-tg-bot/pkg/ydb"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

// type TransCat struct {
// 	Category  sql.NullString `db:"trans_cat"`
// 	Direction sql.NullInt16  `db:"direction"`
// 	ClientID  sql.NullString `db:"client_id"`
// 	Active    sql.NullBool   `db:"active"`
// 	Limit     sql.NullInt64  `db:"trans_limit"`
// }

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
