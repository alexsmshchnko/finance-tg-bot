package repository

import (
	"context"
	"database/sql"
	"finance-tg-bot/internal/entity"
	"finance-tg-bot/pkg/ydb"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

type Repository struct {
	*ydb.Ydb
}

type Reporter interface {
	GetStatementCatTotals(ctx context.Context, p map[string]string) (rres []entity.ReportResult, err error)
}

func (r *Repository) GetStatementCatTotals(ctx context.Context, p map[string]string) (rres []entity.ReportResult, err error) {
	query := `DECLARE $client_id   AS String;
	          DECLARE $datefrom    AS Datetime;
			  DECLARE $dateto      AS Datetime;`
	if _, ok := p["subcat"]; ok {
		query = query + `
select d.trans_cat as trans_cat
      ,"" as comment
      ,case d.direction
         when 0 then sum(d.trans_amount)
         else d.direction * sum(d.trans_amount) 
       end as t_sum
      ,sum(d.trans_amount) as ta, 1 as tp
  from doc d
 inner join client c on (c.id = d.client_id)
 where d.trans_date between $datefrom and $dateto
   and d.trans_amount != 0
   and c.username = $client_id
   and c.is_active
 group by d.trans_cat, d.direction
 union all
select d.trans_cat as trans_cat
	  ,d.comment as comment
	  ,case d.direction
		 when 0 then sum(d.trans_amount)
		 else d.direction * sum(d.trans_amount) 
		 end as t_sum
	  ,sum(d.trans_amount) as ta, 2 as tp
  from doc d
 inner join client c on (c.id = d.client_id)
 where d.trans_date between $datefrom and $dateto
   and d.trans_amount != 0
   and c.username = $client_id
   and c.is_active
 group by d.trans_cat, d.comment, d.direction
 order by trans_cat, tp ASC, ta DESC`
	} else {
		query = query + `
select d.trans_cat as trans_cat
      ,"" as comment
      ,case d.direction
         when 0 then sum(d.trans_amount)
         else d.direction * sum(d.trans_amount) 
       end as t_sum
      ,sum(d.trans_amount) as ta, 1 as tp
  from doc d
 inner join client c on (c.id = d.client_id)
 where d.trans_date between $datefrom and $dateto
   and d.trans_amount != 0
   and c.username = $client_id
   and c.is_active
 group by d.trans_cat, d.direction
union all
select case d.direction
         when -1 then 'Total debit'
         when 0  then 'Total deposit'
         else         'Total credit'
       end as trans_cat
	  ,"" as comment
      ,case d.direction
         when 0 then sum(d.trans_amount)
         else d.direction * sum(d.trans_amount) 
       end as t_sum
      ,sum(d.trans_amount) as ta, 2 as tp
  from doc d
 inner join client c on (c.id = d.client_id)
 where d.trans_date between $datefrom and $dateto
   and d.trans_amount != 0
   and c.username = $client_id
   and c.is_active
 group by d.direction
 order by tp ASC, ta DESC;`
	}

	err = r.Ydb.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		t1, _ := time.Parse("02.01.2006", p["datefrom"])
		t2, _ := time.Parse("02.01.2006", p["dateto"])
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			query,
			table.NewQueryParameters(
				table.ValueParam("$client_id", types.BytesValueFromString(p["username"])),
				table.ValueParam("$datefrom", types.DatetimeValueFromTime(t1)),
				table.ValueParam("$dateto", types.DatetimeValueFromTime(t2)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		if err = res.NextResultSetErr(ctx); err != nil {
			return err
		}
		var (
			rn   sql.NullString
			rc   sql.NullString
			rv   sql.NullInt64
			name string
		)
		for res.NextRow() {
			err = res.ScanNamed(
				named.Optional("trans_cat", &rn),
				named.Optional("comment", &rc),
				named.Optional("t_sum", &rv),
			)
			if err != nil {
				return err
			}
			if _, ok := p["subcat"]; ok {
				if rc.String == "" {
					name = rn.String + ":"
				} else {
					name = rc.String
				}
				rres = append(rres, entity.ReportResult{
					Name: name,
					Val:  int(rv.Int64),
				})
			} else {
				rres = append(rres, entity.ReportResult{
					Name: rn.String,
					Val:  int(rv.Int64),
				})
			}

		}
		return res.Err() // for driver retry if not nil
	})
	return rres, err
}
